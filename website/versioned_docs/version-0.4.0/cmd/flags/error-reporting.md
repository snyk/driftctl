---
id: error-reporting
title: Error Reporting
---

When a crash occurs in driftctl, we do not send any crash reports.

For debugging purposes, you can add `--error-reporting` when running driftctl and crash data will be sent to us via [Sentry](https://sentry.io).

Below is a list of data we retrieve when error reporting is enabled.

- **date**: Event date
- **os name**: Operating System (string, e.g. : "linux | mac | windows")
- **architecture**: Architecture of your CPU (string, e.g. : "amd64 | i389")
- **num_cpu**: Number of cores of your CPU (int, e.g. : 8)
- **release**: driftctl version (string, e.g. : "v0.2.2")
- **server_name**: Your computer hostname (string, e.g. : "yourhostname")
- **runtime version**: Golang version (string, e.g. : "go1.15.2")
- **runtime infos**: Variables go_maxprocs, go_numcgocalls, go_numroutines
- **packages**: Golang used packages and their versions
- **stacktrace**: The error stack

## Example

Below is a full example of a nil pointer crash report.

![Sentry](../../assets/sentry.png)

The RAW stack for this example is:

```console
runtime.errorString: runtime error: invalid memory address or nil pointer dereference
  File "/go/src/app/pkg/parallel_runner.go", line 93, in (*ParallelRunner).Run.func1.1
  File "/go/src/app/pkg/remote/aws/s3_bucket_supplier.go", line 71, in readBucketRegion
  File "/go/src/app/pkg/remote/aws/s3_bucket_inventory_supplier.go", line 42, in (*S3BucketInventorySupplier).Resources
  File "/go/src/app/pkg/scanner.go", line 28, in (*Scanner).Resources.func1
  File "/go/src/app/pkg/parallel_runner.go", line 97, in (*ParallelRunner).Run.func1
```
