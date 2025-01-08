---
title: "Troubleshooting: Getting Started with NodeKit"
sidebar:
  label: "Getting Started with NodeKit"
description: Troubleshooting nodekit installation
---

:::note
This section outlines **common errors encountered** when executing the nodekit installation command.

**If you are looking for the installation instructions instead, they are located [here](/guides/10-getting-started).**
:::

### A nodekit file already exists in the current directory.

If you run the installer command more than once, you will see:

> ERROR: A nodekit file already exists in the current directory. Delete or rename it before installing.

If you want to fetch the latest version of nodekit, you can delete the existing file:

```bash
rm nodekit
```

And then run the [Getting Started](/guides/10-getting-started) command again.

### Command not found: bash

Some versions of Mac OS may not include the required `bash` executable that runs the installer.

If you get an error about `bash` not being available, please install bash on your system manually.

For Mac OS, a popular way to do this is to install [Homebrew](https://brew.sh/) and then install bash using:

```bash
brew install bash
```
