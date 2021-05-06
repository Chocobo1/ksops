# ksops
A Kubernetes KMS provider with [mozilla/sops](https://github.com/mozilla/sops) as the backend

```
   ┌────────────┐                    ┌─────────┐                    ┌────────────┐
   │ Kubernetes │ ─────────────────► │         │ ─────────────────► │            │
   │            │  TCP / UDP / unix  │  ksops  │  TCP / UDP / unix  │mozilla/sops│
   │  kubectl   │ ◄───────────────── │         │ ◄───────────────── │            │
   └────────────┘                    └─────────┘                    └────────────┘
```

# Build

1. Install [Go](https://golang.org/)
2. Download ksops source code
3. Build ksops
   ```shell
   cd <project_directory>
   go build ./cmd/ksops
   # executable `ksops` produced!
   ```

# Example usage

1. Start mozilla/sops and listen on Unix domain sockets at `/tmp/sops.sock`:
   ```shell
   sops keyservice \
     --network unix \
     --address "/tmp/sops.sock"
   ```
2. Start ksops and listen on Unix domain sockets at `/tmp/ksops.sock`: \
   Here we configure ksops to use the PGP key for encryption/decryption.
   Actually mozilla/sops is doing the work, ksops only forwards the encrypt/decrypt request with the proper credentials.
   ```shell
   ./ksops \
     --pgp.key="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4" \
     "unix:/tmp/ksops.sock" \
     "unix:/tmp/sops.sock"
   ```
3. Setup Kubernetes to use a KMS provider (ksops) for data encryption/decryption \
   Read the [documentation](https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/)!

# Synopsis
```shell
Usage:
  ksops [OPTIONS] Address SopsAddress

Age Group Options:
      --age.recipient=     Age recipient

AWS Group Options:
      --aws.arn=           AWS ARN (Amazon Resource Name)
      --aws.role=          AWS IAM role
      --aws.context=       AWS encryption context
      --aws.profile=       AWS profile

Azure Group Options:
      --azure.url=         Azure vault URL
      --azure.key_name=    Azure key name
      --azure.key_version= Azure key version

GCP Group Options:
      --gcp.id=            GCP KMS resource ID

Hashicorp Vault Group Options:
      --vault.address=     Vault address
      --vault.engine_path= Vault transit secrets engine path
      --vault.key=         Vault key

PGP Group Options:
      --pgp.key=           PGP key

Help Options:
  -h, --help               Show this help message

Arguments:
  Address:                 Server listen address. For example: "tcp:127.0.0.1:12345". https://golang.org/pkg/net/#Listen
  SopsAddress:             Sops program keyservice address. https://github.com/grpc/grpc/blob/master/doc/naming.md
```
