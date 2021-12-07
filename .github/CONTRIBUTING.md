# Contributing to driftctl

This document provides guidance on driftctl contribution recommended practices.
It covers what we're looking for in order to help set some expectations and help you get the most out of participation in this project.

To record a bug report, enhancement proposal, or give any other product feedback, please open a GitHub issue using the most appropriate issue template. Please do fill in all of the information the issue templates request, because we've seen from experience that this will maximize the chance that we'll be able to act on your feedback.

## Contributing Fixes

For new contributors we've labeled a few issues with Good First Issue as a nod to issues which will help get you familiar with driftctl development, while also providing an onramp to the codebase itself.

Read the documentation, and don't be afraid to [ask questions](https://discord.gg/NMCBxtD7Nd).

## Proposing a Change

In order to be respectful of the time of community contributors, we aim to discuss potential changes in GitHub issues prior to implementation. That will allow us to give design feedback up front and set expectations about the scope of the change, and, for larger changes, how best to approach the work such that the driftctl team can review it and merge it along with other concurrent work.

If the bug you wish to fix or enhancement you wish to implement isn't already covered by a GitHub issue that contains feedback from the driftctl team, please do start a discussion (either in a new GitHub issue or an existing one, as appropriate) before you invest significant development time. If you mention your intent to implement the change described in your issue, the driftctl team can, as best as possible, prioritize including implementation-related feedback in the subsequent discussion.

Most changes will involve updates to the test suite, and changes to driftctl's documentation. The driftctl team can advise on different testing strategies for specific scenarios, and may ask you to revise the specific phrasing of your proposed documentation prose to match better with the standard "voice" of driftctl's documentation.

### Maintainers

Maintainers are key contributors to our Open Source project.
They contribute their time and expertise and we ask that the community take extra special care to be mindful of this when interacting with them.
There is no expectation on response time for our maintainers; they may be indisposed for prolonged periods of time. Please be patient.

### Pull Request Lifecycle

1. You are welcome to submit a draft pull request for commentary or review before it is fully completed. It's also a good idea to include specific questions or items you'd like feedback on.
2. Once you believe your pull request is ready to be merged you can create your pull request.
3. When time permits driftctl's team members will look over your contribution and either merge, or provide comments letting you know if there is anything left to do. It may take some time for us to respond. We may also have questions that we need answered about the code, either because something doesn't make sense to us or because we want to understand your thought process. **We kindly ask that you do not target specific team members**.
4. If we have requested changes, you can either make those changes or, if you disagree with the suggested changes, we can have a conversation about our reasoning and agree on a path forward. This may be a multi-step process. Our view is that pull requests are a chance to collaborate, and we welcome conversations about how to do things better. It is the contributor's responsibility to address any changes requested. While reviewers are happy to give guidance, **it is unsustainable for us to perform the coding work necessary to get a PR into a mergeable state**.
5. Once all outstanding comments and checklist items have been addressed, your contribution will be merged!
6. In some cases, we might decide that a PR should be closed without merging. We'll make sure to provide clear reasoning when this happens. Following the recommended process above is one of the ways to ensure you don't spend time on a PR we can't or won't merge.

### Getting Your Pull Requests Merged Faster
1. **Well-documented**: Try to explain in the pull request comments what your change does, why you have made the change, and provide instructions for how to produce the new behavior introduced in the pull request. If you can, provide screen captures or terminal output to show what the changes look like. This helps the reviewers understand and test the change.
2. **Small**: Try to only make one change per pull request. If you found two bugs and want to fix them both, that's awesome, but it's still best to submit the fixes as separate pull requests. This makes it much easier for reviewers to keep in their heads all of the implications of individual code changes, and that means the PR takes less effort and energy to merge. In general, the smaller the pull request, the sooner reviewers will be able to make time to review it.
3. **Passing Tests**: Based on how much time we have, we may not review pull requests which aren't passing our tests (look below for advice on how to run unit tests). If you need help figuring out why tests are failing, please feel free to ask, but while we're happy to give guidance it is generally your responsibility to make sure that tests are passing. If your pull request changes an interface or invalidates an assumption that causes a bunch of tests to fail, then you need to fix those tests before we can merge your PR.

If we request changes, try to make those changes in a timely manner. Otherwise, PRs can go stale and be a lot more work for all of us to merge in the future.

Even with everyone making their best effort to be responsive, it can be time-consuming to get a PR merged. It can be frustrating to deal with the back-and-forth as we make sure that we understand the changes fully. Please bear with us, and please know that we appreciate the time and energy you put into the project.

### PR Checks

The following checks run when a PR is opened:

* Tests: tests include unit tests and acceptance tests, and all tests must pass before a PR can be merged.
* Test Coverage Report: We use Codecov to check both overall test coverage, and patch coverage. We are still deciding on the right targets for our code coverage check. A failure in Codecov does not necessarily mean that your PR will not be approved or merged.

## Development Environment

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

## Generated Code

Some files in the driftctl codebase are generated (resources DTO).
Generation is done by an external tool that read schemas from terraform providers and output golang structures with proper tags.
Currently, we have not open-sourced it, but we plan to do it very soon as it will save us lots of time when adding new resources.

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

