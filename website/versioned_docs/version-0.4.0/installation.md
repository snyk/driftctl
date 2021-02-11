---
id: installation
title: Installation
---

driftctl is available on Linux, macOS and Windows.

Binaries are available in the [release page](https://github.com/cloudskiff/driftctl/releases).

## Package manager for macOS

### Homebrew

```shell
$ brew install driftctl
```

### MacPorts

```shell
$ sudo port install driftctl
```

## Docker

```shell
$ docker run -t --rm \
  -v ~/.aws:/home/.aws:ro \
  -v $(pwd):/app:ro \
  -v ~/.driftctl:/home/.driftctl \
  -e AWS_PROFILE=non-default-profile \
  cloudskiff/driftctl scan
```

`-v ~/.aws:/home/.aws:ro` (optionally) mounts your `~/.aws` containing AWS credentials and profile

`-v $(pwd):/app:ro` (optionally) mounts your working dir containing the terraform state

`-v ~/.driftctl:/home/.driftctl` (optionally) prevents driftctl to download the provider at each run

`-e AWS_PROFILE=cloudskiff` (optionally) exports the non-default AWS profile name to use

`cloudskiff/driftctl:<VERSION_TAG>` run a specific driftctl tagged release

## Manual

- **Linux**

This is an example using `curl`. If you don't have `curl`, install it, or use `wget`.

```shell
# x64
$ curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_linux_amd64 -o driftctl

# x86
$ curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_linux_386 -o driftctl
```

Make the binary executable:

```shell
$ chmod +x driftctl
```

Optionally install driftctl to a central location in your `PATH`:

```shell
# use any path that suits you, this is just a standard example. Install sudo if needed.
$ sudo mv driftctl /usr/local/bin/
```

- **macOS**

```shell
# x64
$ curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_darwin_amd64 -o driftctl
```

Make the binary executable:

```shell
$ chmod +x driftctl
```

Optionally install driftctl to a central location in your `PATH`:

```shell
# use any path that suits you, this is just a standard example. Install sudo if needed.
$ sudo mv driftctl /usr/local/bin/
```

- **Windows**

```shell
# x64
$ curl https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_windows_amd64.exe -o driftctl.exe
# x86
$ curl https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_windows_386.exe -o driftctl.exe
```

## Verify digital signatures

Cloudskiff releases are signed using PGP key (ed25519) with ID `ACC776A79C824EBD` and fingerprint `2776 6600 5A7F 01D4 84F6 376D ACC7 76A7 9C82 4EBD`
Our key can be retrieved from common keyservers.

```shell
# Download binary, checksums and signature
$ curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_linux_amd64 -o driftctl_linux_amd64
$ curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_SHA256SUMS -o driftctl_SHA256SUMS
$ curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_SHA256SUMS.gpg -o driftctl_SHA256SUMS.gpg

# Import key
$ gpg --keyserver hkps.pool.sks-keyservers.net --recv-keys 0xACC776A79C824EBD
gpg: key ACC776A79C824EBD: public key "Cloudskiff <security@cloudskiff.com>" imported
gpg: Total number processed: 1
gpg:               imported: 1

# Verify signature (optionally trust the key from gnupg to avoid any warning)
$ gpg --verify driftctl_SHA256SUMS.gpg
gpg: Signature made jeu. 04 févr. 2021 14:58:06 CET
gpg:                using EDDSA key 277666005A7F01D484F6376DACC776A79C824EBD
gpg:                issuer "security@cloudskiff.com"
gpg: Good signature from "Cloudskiff <security@cloudskiff.com>" [ultimate]

# Verify checksum
$ sha256sum --ignore-missing -c driftctl_SHA256SUMS
driftctl_linux_amd64: OK
```

## Run driftctl

Be sure to have [configured](providers/aws/authentication.md) your AWS credentials.

You will need to assign [proper permissions](providers/aws/authentication.md#least-privileged-policy) to allow driftctl to scan your account.

```shell
# With a local state
$ driftctl scan
# Same as
$ driftctl scan --from tfstate://terraform.tfstate

# To specify AWS credentials
$ AWS_ACCESS_KEY_ID=XXX AWS_SECRET_ACCESS_KEY=XXX driftctl scan
# or using a named profile
$ AWS_PROFILE=profile_name driftctl scan

# With state stored on a s3 backend
$ driftctl scan --from tfstate+s3://my-bucket/path/to/state.tfstate

# With multiples states
$ driftctl scan --from tfstate://terraform_S3.tfstate --from tfstate://terraform_VPC.tfstate
```
