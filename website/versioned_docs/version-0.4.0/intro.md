---
id: intro
title: What is driftctl?
slug: /
---

driftctl is CLI tool that measures infrastructure as code coverage, and tracks infrastructure drift.

## Why driftctl?

Infrastructure as code is awesome, but there are too many moving parts: codebase, state file, actual cloud state. Things tend to drift.

Drift can have multiple causes: from developers creating or updating infrastructure through the web console without telling anyone, to uncontrolled updates on the cloud provider side. Handling infrastructure drift vs the codebase can be challenging.

You can't efficiently improve what you don't track. We track coverage for unit tests, why not infrastructure as code coverage?

driftctl tracks how well your IaC codebase covers your cloud configuration. driftctl warns you about drift.

## Features

- **Scan** cloud provider and map resources with IaC code
- Analyze diff, and warn about drift and unwanted unmanaged resources
- Allow users to **ignore** resources
- Multiple output formats

If you want to learn more, find below a good introduction talk:

[Infrastructure drifts aren’t like Pokemon. You can’t catch ”em all.](https://www.youtube.com/watch?v=wDRr2i6XOa0)
