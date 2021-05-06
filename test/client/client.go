package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	k8s "k8s.io/apiserver/pkg/storage/value/encrypt/envelope/v1beta1"
)

//
type ClientOptions struct {
	Address string `description:"K8s KMS provider address. https://github.com/grpc/grpc/blob/master/doc/naming.md"`
}

type CmdOptions struct {
	Client ClientOptions `positional-args:"true" required:"true"`
}

//
const KMSPluginVersion = "v1beta1"

//
func main() {
	// parse cmd arguments
	cmdOptions := CmdOptions{}
	if remainingArgs, err := flags.Parse(&cmdOptions); err != nil {
		if len(remainingArgs) == 0 {
			return
		} else {
			os.Exit(1)
		}
	}

	invokeK8sRPC(cmdOptions.Client.Address, func(ctx *context.Context, client k8s.KeyManagementServiceClient) error {
		request := k8s.VersionRequest{Version: KMSPluginVersion}

		log.Printf("----> Sending RPC Version request: %v", request)
		reply, err := client.Version(*ctx, &request)
		log.Printf("----> Got Version response: %v", reply)

		return err
	})

	invokeK8sRPC(cmdOptions.Client.Address, func(ctx *context.Context, client k8s.KeyManagementServiceClient) error {
		plaintext := []byte("0123456")
		request := k8s.EncryptRequest{
			Version: KMSPluginVersion,
			Plain:   plaintext,
		}

		log.Printf("----> Sending RPC Encrypt request: %v", request)
		reply, err := client.Encrypt(*ctx, &request)
		log.Printf("----> Got Encrypt response: %v", reply)

		return err
	})

	invokeK8sRPC(cmdOptions.Client.Address, func(ctx *context.Context, client k8s.KeyManagementServiceClient) error {
		cipher := []byte("-----BEGIN PGP MESSAGE-----\n\nhF4DNxnzJqb0E2YSAQdA5DwP3Eo70lzyISTArA6Zu/ftIqUkukHyRkL5/1MSWgYw\nWqlCwTTbEcUGwwpbSDhRX7GhjT5070CK1IyCnV5YGJZErqjj5b60k2vBp8ct41Rw\n0ksBI13MaTcKkcYrz+CFGZ4orxtDivEvfBVaCZiQu8y/4sI9z4F4A8Us7t38UBSw\ngmCBeMLY/PrtmRPKZeF7YjgqHQpo2+AmyoeOb4A=\n=0Zr0\n-----END PGP MESSAGE-----\n")
		request := k8s.DecryptRequest{
			Version: KMSPluginVersion,
			Cipher:  cipher,
		}

		log.Printf("----> Sending RPC Decrypt request: %v", request)
		reply, err := client.Decrypt(*ctx, &request)
		log.Printf("----> Got Decrypt response: %v", reply)

		return err
	})
}

//
type KmsProviderRPC = func(*context.Context, k8s.KeyManagementServiceClient) error

func invokeK8sRPC(address string, rpc KmsProviderRPC) error {
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	log.Printf("Connecting to KMS provider on %v", address)
	connection, err := grpc.Dial(address, dialOptions...)
	if err != nil {
		log.Fatalf("Cannot connect to KMS provider: %v", err)
	}
	defer connection.Close()
	log.Printf("Connected to KMS provider on %v", address)

	ctx, cancel := context.WithTimeout(context.Background(), (time.Second * 30))
	defer cancel()

	client := k8s.NewKeyManagementServiceClient(connection)

	return rpc(&ctx, client)
}
