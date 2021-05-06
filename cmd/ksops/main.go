package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"
	sops "go.mozilla.org/sops/v3/keyservice"
	"google.golang.org/grpc"
	k8s "k8s.io/apiserver/pkg/storage/value/encrypt/envelope/v1beta1"
)

// Program properties
const ProgramName = "ksops"
const ProgramVersion = "1.0.0"

// program arg structs
type AgeKey struct {
	Recipient string `long:"recipient" description:"Age recipient"`
}

type AwsKey struct {
	Arn     string            `long:"arn" description:"AWS ARN (Amazon Resource Name)"`
	Role    string            `long:"role" description:"AWS IAM role"`
	Context map[string]string `long:"context" description:"AWS encryption context"`
	Profile string            `long:"profile" description:"AWS profile"`
}

type AzureKey struct {
	Url        string `long:"url" description:"Azure vault URL"`
	KeyName    string `long:"key_name" description:"Azure key name"`
	KeyVersion string `long:"key_version" description:"Azure key version"`
}

type GcpKey struct {
	Id string `long:"id" description:"GCP KMS resource ID"`
}

type HashicorpVaultKey struct {
	Address    string `long:"address" description:"Vault address"`
	EnginePath string `long:"engine_path" description:"Vault transit secrets engine path"`
	Key        string `long:"key" description:"Vault key"`
}

type PgpKey struct {
	Key string `long:"key" description:"PGP key"`
}

type ServerOptions struct {
	Address string `description:"Server listen address. For example: \"tcp:127.0.0.1:12345\". https://golang.org/pkg/net/#Listen"`
}

type SopsOptions struct {
	SopsAddress string `description:"Sops program keyservice address. https://github.com/grpc/grpc/blob/master/doc/naming.md"`
}

type CmdOptions struct {
	Server ServerOptions `positional-args:"true" required:"true"`
	Sops   SopsOptions   `positional-args:"true" required:"true"`

	Age   AgeKey            `group:"Age Group Options" namespace:"age"`
	Aws   AwsKey            `group:"AWS Group Options" namespace:"aws"`
	Azure AzureKey          `group:"Azure Group Options" namespace:"azure"`
	Gcp   GcpKey            `group:"GCP Group Options" namespace:"gcp"`
	Vault HashicorpVaultKey `group:"Hashicorp Vault Group Options" namespace:"vault"`
	Pgp   PgpKey            `group:"PGP Group Options" namespace:"pgp"`
}

// credential type
type CredentialType = int

const (
	CredentialInvalid = iota
	CredentialAge
	CredentialAws
	CredentialAzure
	CredentialGcp
	CredentialHashicorpvault
	CredentialPgp
)

func (options CmdOptions) findCredentialType() CredentialType {
	hasValue := func(str string) bool {
		return len(str) > 0
	}

	if hasValue(options.Age.Recipient) {
		return CredentialAge
	}
	if hasValue(options.Aws.Arn) || hasValue(options.Aws.Role) || (len(options.Aws.Context) > 0) || hasValue(options.Aws.Profile) {
		return CredentialAws
	}
	if hasValue(options.Azure.Url) || hasValue(options.Azure.KeyName) || hasValue(options.Azure.KeyVersion) {
		return CredentialAzure
	}
	if hasValue(options.Gcp.Id) {
		return CredentialGcp
	}
	if hasValue(options.Vault.Address) || hasValue(options.Vault.EnginePath) || hasValue(options.Vault.Key) {
		return CredentialHashicorpvault
	}
	if hasValue(options.Pgp.Key) {
		return CredentialPgp
	}
	return CredentialInvalid
}

// global variable
type Settings struct {
	cmdOptions     CmdOptions
	sopsRequestKey sops.Key
}

func (settings *Settings) setRequestKey() {
	credType := settings.cmdOptions.findCredentialType()

	switch credType {
	case CredentialInvalid:
		log.Fatalf("Missing group option, please provide options from one group (Age, AWS, Azure, GCP, Hashicorp Vault or PGP)")
	case CredentialAge:
		settings.sopsRequestKey = settings.cmdOptions.Age.toSopsKey()
	case CredentialAws:
		settings.sopsRequestKey = settings.cmdOptions.Aws.toSopsKey()
	case CredentialAzure:
		settings.sopsRequestKey = settings.cmdOptions.Azure.toSopsKey()
	case CredentialGcp:
		settings.sopsRequestKey = settings.cmdOptions.Gcp.toSopsKey()
	case CredentialHashicorpvault:
		settings.sopsRequestKey = settings.cmdOptions.Vault.toSopsKey()
	case CredentialPgp:
		settings.sopsRequestKey = settings.cmdOptions.Pgp.toSopsKey()
	default:
		log.Fatal("Missing case: %v", credType)
	}
}

var globalSettings Settings = Settings{}

// functions
func main() {
	// parse cmd arguments
	if remainingArgs, err := flags.Parse(&globalSettings.cmdOptions); err != nil {
		if len(remainingArgs) == 0 {
			return
		} else {
			os.Exit(1)
		}
	}
	globalSettings.setRequestKey()

	// setup server
	listenProtocol, listenAddress, err := parseListenAddress(globalSettings.cmdOptions.Server.Address)
	if err != nil {
		log.Fatalf("Invalid server listen address: %v", err)
	}
	listener, err := net.Listen(listenProtocol, listenAddress)
	if err != nil {
		log.Fatalf("net.Listen() failed: %v", err)
	}
	log.Printf("Listening on %s", globalSettings.cmdOptions.Server.Address)

	server := grpc.NewServer()
	k8s.RegisterKeyManagementServiceServer(server, &K8sServer{})

	// signal handler for cleaning up
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill, syscall.SIGTERM)
	signalHandler := func() {
		caughtsignal := <-signalChannel
		log.Printf("Caught signal: %v. Shutting down.", caughtsignal)
		server.GracefulStop()
		listener.Close()
		os.Exit(0)
	}
	go signalHandler()

	if err := server.Serve(listener); err != nil {
		log.Printf("failed to serve: %v", err)
	}
}

func parseListenAddress(in string) (string, string, error) {
	in = strings.TrimSpace(in)
	strs := strings.SplitN(in, ":", 2)
	if len(strs) != 2 {
		return "", "", fmt.Errorf("Invalid format: %s", in)
	}

	return strs[0], strs[1], nil
}
