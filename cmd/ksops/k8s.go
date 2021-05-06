package main

import (
	"context"
	"log"

	sops "go.mozilla.org/sops/v3/keyservice"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8s "k8s.io/apiserver/pkg/storage/value/encrypt/envelope/v1beta1"
)

type K8sServer struct {
	k8s.UnimplementedKeyManagementServiceServer
}

func (server *K8sServer) Version(k8sCtx context.Context, k8sRequest *k8s.VersionRequest) (*k8s.VersionResponse, error) {
	log.Printf("Version request: version: %v", k8sRequest.GetVersion())
	k8sResponse := k8s.VersionResponse{
		Version:        "v1beta1",
		RuntimeName:    ProgramName,
		RuntimeVersion: ProgramVersion,
	}
	return &k8sResponse, nil
}

func (server *K8sServer) Decrypt(k8sCtx context.Context, k8sRequest *k8s.DecryptRequest) (*k8s.DecryptResponse, error) {
	log.Printf("Got k8s Decrypt request: %v", k8sRequest)

	if !isVersionCompatible(k8sRequest.GetVersion()) {
		log.Printf("Request version incompatible: %v", k8sRequest.GetVersion())
		return nil, status.Error(codes.Unimplemented, "Incompatible request version")
	}

	data, err := invokeSopsRPC(globalSettings.cmdOptions.Sops.SopsAddress, func(sopsCtx *context.Context, client sops.KeyServiceClient) ([]byte, error) {
		reply, err := client.Decrypt(*sopsCtx, &sops.DecryptRequest{
			Key:        &globalSettings.sopsRequestKey,
			Ciphertext: k8sRequest.GetCipher(),
		})
		if err != nil {
			log.Printf("Decrypt at sops failed: %v", err)
		}
		return reply.GetPlaintext(), err
	})
	log.Printf("Decrypt result from sops: %v", data)

	k8sResponse := k8s.DecryptResponse{
		Plain: data,
	}
	log.Printf("Sending Decrypt reply to k8s: %v", k8sResponse)
	return &k8sResponse, err
}

func (server *K8sServer) Encrypt(k8sCtx context.Context, k8sRequest *k8s.EncryptRequest) (*k8s.EncryptResponse, error) {
	log.Printf("Got k8s Encrypt request: %v", k8sRequest)

	if !isVersionCompatible(k8sRequest.GetVersion()) {
		log.Printf("Request version incompatible: %v", k8sRequest.GetVersion())
		return nil, status.Error(codes.Unimplemented, "Incompatible request version")
	}

	data, err := invokeSopsRPC(globalSettings.cmdOptions.Sops.SopsAddress, func(sopsCtx *context.Context, client sops.KeyServiceClient) ([]byte, error) {
		reply, err := client.Encrypt(*sopsCtx, &sops.EncryptRequest{
			Key:       &globalSettings.sopsRequestKey,
			Plaintext: k8sRequest.GetPlain(),
		})
		if err != nil {
			log.Printf("Encrypt at sops failed: %v", err)
		}
		return reply.GetCiphertext(), err
	})
	log.Printf("Encrypt result from sops: %v", data)

	k8sResponse := k8s.EncryptResponse{
		Cipher: data,
	}
	log.Printf("Sending Encrypt reply to k8s: %v", k8sResponse)
	return &k8sResponse, err
}

func isVersionCompatible(version string) bool {
	return version == "v1beta1"
}
