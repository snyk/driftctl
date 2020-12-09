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
  </a><br>
  Measures infrastucture as code coverage, and tracks infrastructure drift.<br>
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
- Multiples output format

## Getting started

### Installation

driftctl is available on Linux, macOS and Windows.

Binaries are available in the [release page](https://github.com/cloudskiff/driftctl/releases).

#### Manual

##### Linux

```bash
# x64
curl https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_linux_amd64 | sudo tee /usr/local/bin/driftctl
# x86
curl https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_linux_386 | sudo tee /usr/local/bin/driftctl
```

##### macOS

```bash
# x64
curl https://github.com/cloudskiff/driftctl/releases/latest/download/driftctl_darwin_amd64 | sudo tee /usr/local/bin/driftctl
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

# With state stored on a s3 backend
$ driftctl scan --from tfstate+s3://my-bucket/path/to/state.tfstate
```

## Documentation & support

- [User guide](doc/README.md)
- [Discord](https://discord.gg/eYGHUa75Q2)

## Contribute

To learn more about compiling driftctl and contributing, please refer to the [contribution guidelines](.github/CONTRIBUTING.md) and [contributing guide](doc/contributing/README.md) for technical details.
