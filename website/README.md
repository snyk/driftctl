# Documentation website for driftctl

https://docs.driftctl.com

This website is built using [Docusaurus 2](https://v2.docusaurus.io/), a static site generator for documentation sites.

## Get Started

1. Make sure all the dependencies for the website are installed:

```shell
# Install dependencies
$ npm install
```

2. Run your dev server:

```shell
# Start the site
$ npm start
```

This starts a server hosting the website locally at http://localhost:3000/. Any changes to the site's source files are reflected after refreshing the page, i.e. the server does not need to be restarted to show changes.

## Documentation

At the root of the repository, the `docs` directory is used to hold documentation written in Markdown. The front matter of new documentation files should at least follow this example:

```yaml
---
id: doc-markdown
title: Markdown File
---

```

For more attributes of the front matter, click [here](https://v2.docusaurus.io/docs/next/api/plugins/@docusaurus/plugin-content-docs#markdown-frontmatter)

The docs are published to `https://docs.driftctl.com/next/` and are considered to be the docs of the next (not yet released) version of driftctl. When a release happens, the docs are copied to the `website/versioned_docs/version-<version>` directory, whose content is published to `https://docs.driftctl.com`.

## Static Build

To create a static build of the website:

```shell
# Build the site
$ npm run build
```

The result will be put in `website/build/`.
