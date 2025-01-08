---
title: Registering your keys online
description: Registering your participation keys online with NodeKit
sidebar:
  order: 40
---

After generating a participation key, you can press `R` to start registering it on the Algorand network.

You can also start this flow by pressing `R` on the [key information screen](/guides/navigating-accounts-and-keys/) shown below.

![](/assets/nodekit-key-info.png)

After you press `R`, you will see a link that you can follow to sign your key registration transaction:

![](/assets/nodekit-keyreg-online.png)

On most terminals, you can hold down Ctrl and click the link, which will open it in your default browser.

If this does not work, copy the link and paste it into your browser.

You will be taken to the Lora Transaction Wizard, where you should see the key information pre-filled:

![](/assets/lora-keyreg.png)

Next you need to:

1. Select `Connect Wallet` on the top right and connect your wallet.

2. Cick the `Send` button on the bottom right. Your wallet should prompt you to sign the key registration transaction

3. Sign the transaction

The transaction will be submitted to the network. If it is accepted, you will see a visual confirmation in Lora similar to the one displayed below:

![](/assets/lora-txn-ok.png)

NodeKit will detect the key registration and take you back to the Key information view:

![](/assets/nodekit-keyreg-success.png)

You can press `ESC` to leave the key information modal.

That's it! Your node is now participating in Algorand consensus. If your account balance is over 30,000 ALGO, it will accumulate rewards for each block it proposes on the Algorand network.
