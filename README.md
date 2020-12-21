<p align="center">
  <img width="201" src="assets/icon.svg" alt="Driftctl"><br>
  <img src="https://circleci.com/gh/cloudskiff/driftctl.svg?style=shield"/>
  <img src="https://goreportcard.com/badge/github.com/cloudskiff/driftctl"/>
  <img src="https://img.shields.io/github/license/cloudskiff/driftctl">
  <img src="https://img.shields.io/github/v/release/cloudskiff/driftctl">
  <img src="https://img.shields.io/github/go-mod/go-version/cloudskiff/driftctl">
  <img src="https://img.shields.io/github/downloads/cloudskiff/driftctl/total.svg"/>
  <a href="https://codecov.io/gh/cloudskiff/driftctl">
    <img src="https://codecov.io/gh/cloudskiff/driftctl/branch/main/graph/badge.svg?token=8C5R02G5S7"/>
  </a>
  <img src="https://img.shields.io/docker/pulls/cloudskiff/driftctl"/>
  <img src="https://img.shields.io/microbadger/layers/cloudskiff/driftctl"/>
  <img src="https://img.shields.io/docker/image-size/cloudskiff/driftctl"/>

<br>
  Measures infrastructure as code coverage, and tracks infrastructure drift.<br>
  :warning: <strong>This tool is still in beta state and will evolve in the future with potential breaking changes</strong> :warning:
</p>

## Why ?

Infrastructure as code is awesome, but there are too many moving parts: codebase, state file, actual cloud state. Things tend to drift.

Drift can have multiple causes: from developers creating or updating infrastructure through the web console without telling anyone, to uncontrolled updates on the cloud provider side. Handling infrastructure drift vs the codebase can be challenging.

You can't efficiently improve what you don't track. We track coverage for unit tests, why not infrastructure as code coverage?

driftctl tracks how well your IaC codebase covers your cloud configuration. driftctl warns you about drift.

## Features

- **Scan** cloud provider and map resources with IaC code
- Analyze diff, and warn about drift and unwanted unmanaged resources
- Allow users to **ignore** resources
- Multiple output formats

## Getting started

### Installation

driftctl is available on Linux, macOS and Windows.

Binaries are available in the [release page](https://github.com/cloudskiff/driftctl/releases).

#### Docker

```bash
docker run \
  -v ~/.aws:/app/.aws:ro \
  -v $(pwd)/terraform.tfstate:/app/terraform.tfstate:ro \
  -v ~/.driftctl:/app/.driftctl \
  -e AWS_REGION=us-east-1 \
  -e AWS_PROFILE=cloudskiff \
  cloudskiff/driftctl:v0.1.1 scan ## with the same option as the binary version
```

`-v ~/.aws:/app/.aws:ro` mount your .aws containing credentials and profile

`-v $(pwd)/terraform.tfstate:/app/terraform.tfstate:ro` mount your terraform state

`-v ~/.driftctl:/app/.driftctl` to prevent driftctl downloading the provider at each run, mount a directory to persist it

`-e AWS_PROFILE=cloudskiff` export the profile to use in your AWS config

`-e AWS_REGION=us-east-1` export the specific AWS region you want to use

`cloudskiff/driftctl:<VERSION_TAG>` run a specific driftctl tagged release

#### Manual

##### Linux

This is an example using `curl`. If you don't have `curl`, install it, or use `wget`.

```bash
# x64
curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_linux_amd64 -o driftctl

# x86
curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_linux_386 -o driftctl
```

Make the binary executable:

```bash
chmod +x driftctl
```

Optionally install driftctl to a central location in your `PATH`:

```bash
# use any path that suits you, this is just a standard example. Install sudo if needed. 
sudo mv driftctl /usr/local/bin/
```

##### macOS

```bash
# x64
curl -L https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_darwin_amd64 -o driftctl
```

Make the binary executable:

```bash
chmod +x driftctl
```

Optionally install driftctl to a central location in your `PATH`:

```bash
# use any path that suits you, this is just a standard example. Install sudo if needed. 
sudo mv driftctl /usr/local/bin/
```

##### Windows

```bash
# x64
curl https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_windows_amd64.exe -o driftctl.exe
# x86
curl https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_windows_386.exe -o driftctl.exe
```

### Run

Be sure to have [configured](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) your AWS credentials.

You will need to assign [proper permissions](doc/cmd/scan/supported_resources/aws.md#least-privileged-policy) to allow driftctl to scan your account.

```bash
# With a local state
$ driftctl scan
# Same as
$ driftctl scan --from tfstate://terraform.tfstate

# To specify AWS credentials
$ AWS_ACCESS_KEY_ID=XXX AWS_SECRET_ACCESS_KEY=XXX driftctl scan
# or using a profile
$ AWS_PROFILE=profile_name driftctl scan

# With state stored on a s3 backend
$ driftctl scan --from tfstate+s3://my-bucket/path/to/state.tfstate
```
## Documentation & support

- [User guide](doc/README.md)
- [Discord](https://discord.gg/eYGHUa75Q2)

## Contribute

To learn more about compiling driftctl and contributing, please refer to the [contribution guidelines](.github/CONTRIBUTING.md) and [contributing guide](doc/contributing/README.md) for technical details.
