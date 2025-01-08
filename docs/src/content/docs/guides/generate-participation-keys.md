---
title: Generating participation keys
description: Generating participation keys with NodeKit
sidebar:
  order: 30
---

If it is not running already, start NodeKit with this command:

```bash
./nodekit
```

After your node has fully synced with the network, you will see a green `RUNNING` label at the top:

![](/assets/nodekit-state-running.png)

You will only be able to generate participation keys after your node is in a `RUNNING` state

## Generate participation keys

Press the `G` key to start generating participation keys.

NodeKit will ask the account address that will be participating in consensus. Enter your account address and press `ENTER`.

![](/assets/nodekit-partkey-gen-1.png)

## Select participation key duration

NodeKit will ask the number of days that the participation keys will be valid for:

![](/assets/nodekit-partkey-gen-2.png)

You can press the `S` key to cycle through duration modes in days / months / rounds.

The longer your duration, the longer the participation key generation step will take to complete.

## Key generation

After you have selected your key duration, nodekit will instruct your node to generate participation keys.

The time required for this step will depend on your participation key duration. As an indicative wait time, a 30-day participation key should take between 4-6 minutes to generate.

![](/assets/nodekit-partkey-gen-3.png)

## Done

When your participation keys are ready, nodekit will display the key information as shown below.

![](/assets/nodekit-partkey-gen-4.png)

You are now one step away from participating in Algorand consensus!

As the on screen message indicates, you can press `R` to start [Registering your keys](/guides/register-online).

:::note
**Did you encounted any errors?**
Check out the [Troubleshooting: Generating participation keys](/troubleshooting#generating-participation-keys) section.
:::
