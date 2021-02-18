# User guide

### Global flags

#### Version check

By default, driftctl checks for a new version remotely. To disable this behavior, either use the flag `--no-version-check` or define the environment variable `DCTL_NO_VERSION_CHECK`.

#### Error reporting

When a crash occurs in driftctl, we do not send any crash reports.
For debugging purposes, you can add `--error-reporting` when running driftctl and crash data will be sent to us via [Sentry](https://sentry.io)
Details of reported data can be found [here](./cmd/flags/error-reporting.md)

#### Log level

By default driftctl logger only displays warning and error messages. You can set `LOG_LEVEL` environment variable to change the default level.
Valid values are : trace,debug,info,warn,error,fatal,panic.

**Note:** In trace level, terraform provider logs will be shown.

Example

```shell
$ LOG_LEVEL=debug driftctl scan
DEBU[0000] New provider library created
DEBU[0000] Found existing provider path=/home/driftctl/.driftctl/plugins/linux_amd64/terraform-provider-aws_v3.19.0_x5
DEBU[0000] Starting gRPC client alias=us-east-1
DEBU[0001] New gRPC client started alias=us-east-1
...
```

### Usage

- Commands
  - Scan
    - [Output format](cmd/scan/output.md)
    - [Filtering resources](cmd/scan/filter.md)
    - [Supported remotes](cmd/scan/supported_resources/README.md)
    - [Iac sources](cmd/scan/iac_source.md)
  - [Completion](cmd/completion/script.md)

## Issues

- [Known Issues & Limitations](LIMITATIONS.md)
