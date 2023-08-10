## This project is now in maintenance mode. We cannot promise to review contributions. Please feel free to fork the project to apply any changes you might want to make.











### Build

If you wish to work on the driftctl CLI source code, you'll first need to install the Go compiler and the version control system Git.

At this time the driftctl development environment is targeting only Linux Mac OS X and Windows systems.

Refer to the file [`.go-version`](https://github.com/cloudskiff/driftctl/blob/master/.go-version) to see which version of Go driftctl is currently built with. Other versions will often work, but if you run into any build or testing problems please try with the specific Go version indicated. You can optionally simplify the installation of multiple specific versions of Go on your system by installing [`goenv`](https://github.com/syndbg/goenv), which reads `.go-version` and automatically selects the correct Go version.

Use Git to clone this repository into a location of your choice. driftctl is using [Go Modules](https://blog.golang.org/using-go-modules), and so you should *not* clone it inside your `GOPATH`.

Switch into the root directory of the cloned repository and build driftctl using GNU Make:

```shell script
make build
```

The first time you run the `make build` command, the build script will download any library dependencies that you don't already have in your Go modules cache.
Subsequent builds will be faster because these dependencies will already be available on your local disk.

Once the compilation process succeeds, you can find a `driftctl_$os_$arch` executable in the `bin/` directory.

**Note**: driftctl uses an `.editorconfig` file to normalize indentation stuff and other common guidelines.
We kindly ask you to use an editor that supports it or at least configure your editor parameters according to our guidelines.
Working together with the same guidelines saves us a lot of brainwork during code review and could avoid some conflict.

### Unit test

If you are planning to make changes to the driftctl source code, you should run the unit test suite before you start to make sure everything is initially passing:

```shell script
go test ./...
```

As you make your changes, you can re-run the above command to ensure that the tests are *still* passing. If you are working only on a specific Go package, you can speed up your testing cycle by testing only that single package, or packages under a particular package prefix:

```shell script
go test ./pkg/iac/...
```

For more details on testing, check the [contributing guide](../docs/testing.md).

### Acceptance Tests: Testing interactions with external services

driftctl's unit test suite is self-contained, using mocks and local files to help ensure that it can run offline and is unlikely to be broken by changes made to or coming from outside systems.

There are some optional tests in the driftctl CLI codebase that *do* interact with external services, which we collectively refer to as "acceptance tests".
You can enable these by setting the environment variable `DRIFTCTL_ACC=true` when running the tests.
We recommend focusing only on the specific package you are working on when enabling acceptance tests, both because it can help the test run to complete faster and because you are less likely to encounter failures due to drift in systems unrelated to your current goal:

Because the acceptance tests depend on services outside of the driftctl codebase, and because the acceptance tests are usually used only when making changes to the systems they cover, it is common and expected that drift in those external systems will cause test failures.
Because of this, prior to working on a system covered by acceptance tests it's important to run the existing tests for that system in an *unchanged* work tree first and respond to any test failures that preexist, to avoid misinterpreting such failures as bugs in your new changes.

More details on acceptance on the [contributing guide](../docs/README.md)

## External Dependencies

Terraform uses Go Modules for dependency management.

Our dependency licensing policy for driftctl excludes proprietary licenses and "copyleft"-style licenses.
We will consider other open source licenses in similar spirit to those three, but if you plan to include
such a dependency in a contribution we'd recommend opening a GitHub issue first to discuss what you intend
to implement and what dependencies it will require so that the driftctl team can review the relevant licenses
to for whether they meet our licensing needs.

If you need to add a new dependency to driftctl or update the selected version for an existing one, use `go get` from the root of the driftctl repository as follows:

```shell script
go get github.com/hashicorp/terraform@13.0.0
```

This command will download the requested version (13.0.0 in the above example) and record that version selection in the `go.mod` file.
It will also record checksums for the module in the `go.sum`.

To complete the dependency change, clean up any redundancy in the module metadata files by running:

```shell script
make go.mod
```

To ensure that the upgrade has worked correctly, be sure to run the unit test suite at least once.

Because dependency changes affect a shared, top-level file, they are more likely than some other change types to become conflicted with other proposed changes during the code review process.
For that reason, and to make dependency changes more visible in the change history, we prefer to record dependency changes as separate commits that include only the results of the above commands and the minimal set of changes to driftctl's own code for compatibility with the new version:

```
git add go.mod go.sum
git commit -m "go get github.com/hashicorp/terraform@13.0.0"
```

You can then make use of the new or updated dependency in code added in subsequent commits.

