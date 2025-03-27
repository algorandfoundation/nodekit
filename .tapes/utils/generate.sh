#!/usr/bin/env bash

set -e

# Run the installer script
vhs ./src/install.tape

# Install Algod
rm nodekit
nodekit install
nodekit stop

# Configure fnet

# Update Algod Binary
sudo ./utils/update_binary.sh
echo '{"DNSBootstrapID": "<network>.algorand.green"}' | sudo tee /var/lib/algorand/config.json
./utils/get_genesis.sh | sudo tee /var/lib/algorand/genesis.json

# Import test account
goal wallet new deployer --non-interactive -d /var/lib/algorand
goal account import -m "$DEPLOYER_MNEMONIC" -d /var/lib/algorand

# Start fnet Algod
nodekit start

# Run fast-catchup
vhs ./src/lagging.tape

./utils/wait_sync.sh


vhs ./src/create.tape

vhs ./src/online.tape | tee &

sleep 2
goal account changeonlinestatus --address "$DEPLOYER_ADDR" -d /var/lib/algorand

wait

vhs ./src/offline.tape | tee &

sleep 2
goal account changeonlinestatus --online=false  --address "$DEPLOYER_ADDR" -d /var/lib/algorand

wait

vhs ./src/delete.tape