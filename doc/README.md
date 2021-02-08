# User guide

### Global flags

#### Version check

By default, driftctl checks for a new version remotely. To disable this behavior, either use the flag `--no-version-check` or define the environment variable `DCTL_NO_VERSION_CHECK`.

#### Error reporting

When a crash occurs in driftctl, we do not send any crash reports.
For debugging purposes, you can add `--error-reporting` when running driftctl and crash data will be sent to us via [Sentry](https://sentry.io)
Details of reported data can be found [here](./cmd/flags/error-reporting.md)

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