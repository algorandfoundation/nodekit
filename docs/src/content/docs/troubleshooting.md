---
title: "Troubleshooting"
description: Troubleshooting NodeKit installation
next:
  label: "NodeKit Reference"
  link: /reference/nodekit
---

This page contains troubleshooting tips for common issues that you might run into. For additional support, please visit our [Discord channel](https://discord.com/channels/491256308461207573/807825288666939482).

## Getting started with NodeKit

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

## Bootstrapping the Algorand node

:::note
This section outlines **common errors encountered** during the "bootstrap" node installation process.

**If you are looking for the instructions instead, they are located [here](/guides/20-bootstrap).**
:::

### Asking for password

The installer will ask for your user password during this process. This is required by your operating system in order to install new software.

Enter your operating system user password when you see this prompt:

```
[sudo] password for user:
```

### Fast catchup is taking too long to complete

If the fast catchup process is taking too long to complete, check the following:

#### Status is in FAST-CATCHUP

The colored status at the top of Nodekit should be in a yellow `FAST-CATCHUP` state:

![](/assets/nodekit-state-fast-catchup.png)

After fast-catchup completes successfully, it is normal for a node to be in a `SYNCING` state for a few minutes:

![](/assets/nodekit-state-syncing.png)

During this the `Latest Round` number should be increasing rapidly.

If there is no progress for a while, or the Latest Round value is smaller than `46000000` (46 million) then you should start fast-catchup again:

#### Fast Catchup was interrupted or did not complete

To start the fast catchup process again, exit the Nodekit user interface by pressing `Q` and enter the following command:

```bash
./nodekit catchup
```

You can enter nodekit again by running:

```bash
./nodekit
```

When the syncing process completes, the status will display `RUNNING`:

![](/assets/nodekit-state-running.png)

### Your hardware meets the minimum specs

TODO SSD

### Your network connection meets the minimum specs

TODO

## Generating participation keys

:::note
This section outlines **common errors encountered** during the participation key generation step on Nodekit.

**If you are looking for the instructions instead, they are located [here](/guides/30-generate-participation-keys).**
:::

### Failed to get participation keys

Occasionally, Nodekit may fall out of sync with the Algorand node while waiting for the participation keys to be generated. In this case this error message will be shown:

You can:

- wait for the participation keys to appear in the Accounts list
- try to generate a participation key again
  - If the key generation process is still running on the node, you will see a ["Participation key generation already in progress"](#participation-key-generation-already-in-progress) error

![](/assets/nodekit-error-keygen-failed.png)

### Participation key generation already in progress

This error means that there is a participation key already being generated on the Algorand node.

You can wait a few minutes, and the key will automatically appear in the Accounts list when it is ready.

![](/assets/nodekit-error-keygen-already.png)
