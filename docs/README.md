# Developer guide

This directory contains some documentation about the driftctl codebase, aimed at readers who are interested in making code contributions.

- [Add new remote provider](new-remote-provider.md)
- [Add new resources](new-resource.md)
- [Testing](testing.md)

## Core concepts

driftctl uses Terraform providers besides cloud providers SDK to retrieve data.

Resource listing is done using cloud providers SDK. Resource details retrieval is done by calling terraform providers with gRPC.

## Terminology

- `Remote` is a representation of a cloud provider
- `Resource` is an abstract representation of a cloud provider resource (e.g. S3 bucket, EC2 instance, etc ...)
- `Enumerator` is used to list resources of a given type from a given remote and return a resource list, it should exist only one Enumerator per resource
