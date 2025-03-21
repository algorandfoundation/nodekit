#!/usr/bin/env bash

set -e

LOGPFX=$(basename "$0"):


vhs bootstrap.tape

mkdir -p bin
touch bin/algod
touch bin/goal

sudo ./utils/update_binary.sh
echo '{"DNSBootstrapID": "<network>.algorand.green"}' | sudo tee /var/lib/algorand/config.json
sudo chmod +x bin/algod
sudo bin/algod -d /var/lib/algorand/ &

sleep 5
vhs lagging.tape

sleep 10
vhs catchup.tape
