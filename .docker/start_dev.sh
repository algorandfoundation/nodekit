#!/usr/bin/env bash

set -e

if [ "$DEBUG" = "1" ]; then
    set -x
fi

if [ "$ALGORAND_DATA" != "/algod/data" ]; then
    echo "Do not override 'ALGORAND_DATA' environment variable."
    exit 1
fi

# Configure the participation node
if [ -d "$ALGORAND_DATA" ]; then
    if [ -f "$ALGORAND_DATA/genesis.json" ]; then
        if [ "$TOKEN" != "" ]; then
            echo "$TOKEN" >"$EMPTY_DATA/algod.token"
        fi
        if [ "$ADMIN_TOKEN" != "" ]; then
            echo "$ADMIN_TOKEN" >"$EMPTY_DATA/algod.admin.token"
        fi
        algod -o -d "$ALGORAND_DATA" -l "0.0.0.0:8080"
    else
        sed -i "s/NUM_ROUNDS/${NUM_ROUNDS:-30000}/" "/node/run/template.json"
        sed -i "s/\"NetworkName\": \"\"/\"NetworkName\": \"hack-tui\"/" "/node/run/template.json"
        goal network create --noclean -n tuinet -r "${ALGORAND_DATA}/.." -t "/node/run/template.json"

        # Cycle Network
        goal network start -r "${ALGORAND_DATA}/.."
        goal node stop

        # Update Tokens
        if [ "$TOKEN" != "" ]; then
            echo "$TOKEN" >"$ALGORAND_DATA/algod.token"
        fi
        if [ "$ADMIN_TOKEN" != "" ]; then
            echo "$ADMIN_TOKEN" >"$ALGORAND_DATA/algod.admin.token"
        fi
        # Import wallet
        goal account import -m "artefact exist coil life turtle edge edge inside punch glance recycle teach melody diet method pause slam dumb race interest amused side learn able heavy"

        algod -o -d "$ALGORAND_DATA" -l "0.0.0.0:8080"
    fi

else
    echo "$ALGORAND_DATA" does not exist
    exit 1
fi
