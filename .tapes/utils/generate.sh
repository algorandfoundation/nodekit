#!/usr/bin/env bash

set -e

# Run the installer script
vhs ./src/install.tape

# Install Algod
nodekit install
nodekit stop
rm -rf /var/lib/var/algorand

# Configure fnet

# Update Algod Binary
sudo ./utils/update_binary.sh
echo '{"DNSBootstrapID": "<network>.algorand.green"}' | sudo tee /var/lib/algorand/config.json
./utils/get_genesis.sh | sudo tee /var/lib/algorand/genesis.json

# Start fnet Algod
sudo algod -d /var/lib/algorand/ &

# Run fast-catchup
vhs ./src/lagging.tape

./utils/wait_sync.sh

vhs ./src/create.tape

vhs ./src/online.tape