---
title: Bootstrapping the Algorand node
description: Installing and bootstrapping the Algorand node
sidebar:
  order: 20
---

## How to start the process

The bootstrap process is automatically started after following the [NodeKit installation instructions](/guides/getting-started), but it can also be triggered manually by running this command:

```bash
./nodekit bootstrap
```

## Prompts: Installation and Fast-catchup

`nodekit bootstrap` will check to see if there is a node already installed.

If there is none, it will ask if you want to start a node installation:

> Installing A Node
>
> It looks like you're running this for the first time. Would you like to install a node? (y/n)

You can respond by pressing the `Y` or `N` keys of your keyboard, for `Yes` and `No` respectively.

---

It will then ask if you want to perform a "fast-catchup" with the network:

> Regular sync with the network usually takes multiple days to weeks. You can optionally perform fast-catchup to sync in 30-60 minutes instead.
>
> Would you like to preform a fast-catchup after installation? (y/n)

Fast-catchup saves a lot of time, so we recommend responding Yes.

---

Assuming you have responded "Yes" to the node install prompt, you will now be prompted for your user password:

```
WARN (You may be prompted for your password)
INFO Installing Algod on Linux
INFO Installing with apt-get
[sudo] password for user:
```

Your operating system requires this to allow NodeKit to install the Algorand node software. Enter your user password and press ENTER to proceed.

## Installation

After you enter your password, you can now sit back and wait until your Algorand node is installed and syncs with the network.

The installation phase should only take a few minutes. Your terminal will look like this during the installation phase:

![Screenshot of first phase of "nodekit bootstrap" process](/assets/nodekit-bootstrap.png)

## Fast catchup

After installation is complete, NodeKit will automatically start the NodeKit user interface.

This will display the progress of catching up to the latest state of the Algorand network:

![Screenshot of second phase of "nodekit bootstrap" process - fast catchup](/assets/nodekit-fast-catchup.png)

This process usually takes between 30-60 minutes, depending on your hardware and network connection.

When the process is done, the Fast Catchup status information will disappear and the yellow `FAST-CATCHUP` status at the top will change to a green `RUNNING` state.

![](/assets/nodekit-state-running.png)

:::note
**Did you encounted any errors?**
Check out the [Troubleshooting: Installing and bootstrapping the Algorand node](/troubleshooting#bootstrapping-the-algorand-node) section.
:::
