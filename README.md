<p align="center">
  <img width="200" src="https://docs.driftctl.com/img/driftctl_dark.svg" alt="driftctl">
</p>

<p align="center">
  <img src="https://circleci.com/gh/cloudskiff/driftctl.svg?style=shield"/>
  <img src="https://goreportcard.com/badge/github.com/cloudskiff/driftctl"/>
  <img src="https://img.shields.io/github/license/cloudskiff/driftctl">
  <img src="https://img.shields.io/github/v/release/cloudskiff/driftctl">
  <img src="https://img.shields.io/github/go-mod/go-version/cloudskiff/driftctl">
  <img src="https://img.shields.io/github/downloads/cloudskiff/driftctl/total.svg"/>
  <img src="https://img.shields.io/bintray/dt/homebrew/bottles/driftctl?label=homebrew"/>
  <a href="https://codecov.io/gh/cloudskiff/driftctl">
    <img src="https://codecov.io/gh/cloudskiff/driftctl/branch/main/graph/badge.svg?token=8C5R02G5S7"/>
  </a>
  <img src="https://img.shields.io/docker/pulls/cloudskiff/driftctl"/>
  <img src="https://img.shields.io/microbadger/layers/cloudskiff/driftctl"/>
  <img src="https://img.shields.io/docker/image-size/cloudskiff/driftctl"/>
  <a href="https://discord.gg/NMCBxtD7Nd">
    <img src="https://img.shields.io/discord/783720783469871124?color=%237289da&label=discord&logo=discord"/>
  </a>
</p>

<p align="center">
  Measures infrastructure as code coverage, and tracks infrastructure drift.<br>
  <strong>IaC:</strong> Terraform, <strong>Cloud providers:</strong> AWS, GitHub (Azure and GCP on the roadmap for 2021).<br>
  :warning: <strong>This tool is still in beta state and will evolve in the future with potential breaking changes</strong> :warning:
</p>

## Why driftctl ?

Infrastructure as code is awesome, but there are too many moving parts: codebase, state file, actual cloud state. Things tend to drift.

Drift can have multiple causes: from developers creating or updating infrastructure through the web console without telling anyone, to uncontrolled updates on the cloud provider side. Handling infrastructure drift vs the codebase can be challenging.

You can't efficiently improve what you don't track. We track coverage for unit tests, why not infrastructure as code coverage?

driftctl tracks how well your IaC codebase covers your cloud configuration. driftctl warns you about drift.

## Features

- **Scan** cloud provider and map resources with IaC code
- Analyze diffs, and warn about drift and unwanted unmanaged resources
- Allow users to **ignore** resources
- Multiple output formats

---

**[Get Started](https://driftctl.com/product/quick-tutorial/)**

**[Documentation](https://docs.driftctl.com)**

**[Discord](https://discord.gg/NMCBxtD7Nd)**

---

## Contribute

To learn more about compiling driftctl and contributing, please refer to the [contribution guidelines](.github/CONTRIBUTING.md) and the [contributing guide](docs/README.md) for technical details.

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification and is brought to you by these [awesome contributors](CONTRIBUTORS.md).

Build with ❤️️ from 🇫🇷 🇯🇵 🇬🇷 🇸🇪 🇺🇸
