---
title: Getting Started with NodeKit
description: Installing NodeKit
sidebar:
  order: 10
---

Welcome to NodeKit, your command-line one-stop-shop for Algorand node running.

NodeKit can help you with:

- Installing and configuring an Algorand node
- Syncing your node with the latest state of the network
- Managing consensus participation keys
- Monitoring your node and the network

To get started with NodeKit, copy-paste this command in your terminal:

```bash
wget -qO- https://nodekit.run/install.sh | bash
```

This will detect your operating system and download the appropriate NodeKit executable to your local directory.

It will then immediately start the bootstrap process to get your Algorand node up and running:

![Screenshot of a typical NodeKit installation process](/assets/nodekit-install.png)

:::note
**Did you encounter any errors?**
Check out the [Troubleshooting: Getting Started](/troubleshooting#getting-started-with-nodekit) section.
:::

Otherwise, read on for guidance on the bootstrapping process.
