# Developer guide

- [How to add new resource](./adding_a_new_resource.md)
- [Tests](./tests.md)

## Core concepts

Driftctl uses Terraform providers besides cloud providers SDK to retrieve data.
Resource listing is done using cloud providers SDK, then resource details retrieval is done by calling the terraform provider with gRPC.

## Terminology

- `Scanner` Scanner is used to scan multiples cloud providers and return a set of resources. It calls every declared `Supplier`.
- `Remote` A remote is a representation of a cloud provider
- `Resource` A resource is an abstract representation of a cloud provider resource (e.g. S3 bucket, EC2 instance, etc ...)
- `ResourceSupplier` It should exist only one ResourceSupplier per resource. A ResourceSupplier is used to list resources from a given type on a given remote and return a resource list.
