#!/usr/bin/env bash

set -e

if [ "$DEBUG" = "1" ]; then
  set -x
fi

if [ "$ALGORAND_DATA" != "/algod/data" ]; then
  echo "Do not override 'ALGORAND_DATA' environment variable."
  exit 1
fi

# To allow mounting the data directory we need to change permissions
# to our algorand user. The script is initially run as the root user
# in order to change permissions, afterwards the script is re-launched
# as the algorand user.
if [ "$(id -u)" = '0' ]; then
  chown -R algorand:algorand "$ALGORAND_DATA"
  exec su -p -c "$(readlink -f $0) $@" algorand
fi


# Configure the participation node
if [ -d "$ALGORAND_DATA" ]; then
  if [ "$TOKEN" != "" ]; then
      echo "$TOKEN" > "$ALGORAND_DATA/algod.token"
  fi
  if [ "$ADMIN_TOKEN" != "" ]; then
    echo "$ADMIN_TOKEN" > "$ALGORAND_DATA/algod.admin.token"
  fi
  cd "$ALGORAND_DATA"
  cp "/node/run/genesis/testnet/genesis.json" genesis.json
  algocfg profile set --yes -d "$ALGORAND_DATA" "participation"
  algod -o -d $ALGORAND_DATA -l "0.0.0.0:8080"
else
  echo "$ALGORAND_DATA" does not exist
  exit 1
fi