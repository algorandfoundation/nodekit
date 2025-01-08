---
title: "Installation/Bootstrapping"
description: Troubleshooting nodekit bootstrap
---

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
