package main

import (
	"context"
	"log"
	"time"

	sops "go.mozilla.org/sops/v3/keyservice"
	"google.golang.org/grpc"
)

func (in AgeKey) toSopsKey() sops.Key {
	return sops.Key{
		KeyType: &sops.Key_AgeKey{&sops.AgeKey{
			Recipient: in.Recipient,
		}},
	}
}

func (in AwsKey) toSopsKey() sops.Key {
	return sops.Key{
		KeyType: &sops.Key_KmsKey{&sops.KmsKey{
			Arn:        in.Arn,
			Role:       in.Role,
			Context:    in.Context,
			AwsProfile: in.Profile,
		}},
	}
}

func (in AzureKey) toSopsKey() sops.Key {
	return sops.Key{
		KeyType: &sops.Key_AzureKeyvaultKey{&sops.AzureKeyVaultKey{
			VaultUrl: in.Url,
			Name:     in.KeyName,
			Version:  in.KeyVersion,
		}},
	}
}

func (in GcpKey) toSopsKey() sops.Key {
	return sops.Key{
		KeyType: &sops.Key_GcpKmsKey{&sops.GcpKmsKey{
			ResourceId: in.Id,
		}},
	}
}

func (in HashicorpVaultKey) toSopsKey() sops.Key {
	return sops.Key{
		KeyType: &sops.Key_VaultKey{&sops.VaultKey{
			VaultAddress: in.Address,
			EnginePath:   in.EnginePath,
			KeyName:      in.Key,
		}},
	}
}

func (in PgpKey) toSopsKey() sops.Key {
	return sops.Key{
		KeyType: &sops.Key_PgpKey{&sops.PgpKey{
			Fingerprint: in.Key,
		}},
	}
}

//
type SopsRPC = func(*context.Context, sops.KeyServiceClient) ([]byte, error)

func invokeSopsRPC(address string, rpc SopsRPC) ([]byte, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	log.Printf("Connecting to sops keyservice on %v", address)
	connection, err := grpc.Dial(address, dialOptions...)
	if err != nil {
		log.Printf("Cannot connect to sops keyservice: %v", err)
	}
	defer connection.Close()
	log.Printf("Connected to sops keyservice on %v", address)

	ctx, cancel := context.WithTimeout(context.Background(), (time.Second * 30))
	defer cancel()

	client := sops.NewKeyServiceClient(connection)

	reply, err := rpc(&ctx, client)
	return reply, err
}
