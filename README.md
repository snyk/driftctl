# Cloudskiff
Cloudskiff measures infrastucture as code coverage, and tracks cloud configuration drift.

## Why ?

Infrastructure as code is awesome, but there are too many moving parts: codebase, state file, actual cloud configuration. Things tend to drift.

Drift can have multiple causes: from developers creating or updating infrastructure through the web console without telling anyone, to uncontrolled updates on the cloud provider side. Handling infrastructure drift vs the codebase can be challenging.

You can't efficiently improve what you don't track. **We track coverage for unit tests, why not infrastructure as code coverage?**

Cloudskiff tracks how well your IaC codebase covers your cloud configuration. Cloudskiff warns you about drift.


## What does it do?
- **Scan** cloud provider and map ressources with IaC code
- **Output coverage** in standard format
- **Integrate into CI** flow to expose coverage to everyone
- Allow users to **ignore** resources
- **Analyze diff** between consecutive runs, and warn about drift and unwanted unmanaged ressources

## Getting started

### Installation

#### Docker

```
docker run -it cloudskiff/cloudskiff
```

#### Automated installation

```
curl https://cloudskiff.com/install.sh | sudo bash
```

#### Manual installation

```
curl https://xxxx/clouskiff_amd64_linux -O /usb/bin/cloudskiff
```

### Run coverage


```
$ export AWS_ACCESS_KEY_ID=XXXX
$ export AWS_SECRET_ACCESS_KEY=XXXX
$ export AWS_DEFAULT_REGION=eu-west-1
$ cloudskiff run coverage
```

### Run diff from previous run

```
$ export AWS_ACCESS_KEY_ID=XXXX
$ export AWS_SECRET_ACCESS_KEY=XXXX
$ export AWS_DEFAULT_REGION=eu-west-1
$ cloudskiff run diff --previous-state=/my-previous-state.csstate > diff
$ cat diff | mail -s devops@foobar.com
```

## CI Integrations

### Circle CI

```yaml
version: 2.1
jobs:
  tests:
    steps:
      - run:
          name: Run cloudskiff code coverage
          command: cloudskiff run coverage
workflows:
  push:
    jobs:
      - tests:
```

## Roadmap

- More ressources support
- More cloud provider support
- ...

