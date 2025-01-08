---
title: "Participation key generation"
description: Troubleshooting participation key generation with Nodekit
---

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
