#!/usr/bin/env bash

set -e

if [ "$DEBUG" = "1" ]; then
  set -x
fi

if [ "$ALGORAND_DATA" != "/algod/data" ]; then
  echo "Do not override 'ALGORAND_DATA' environment variable."
  exit 1
fi

EMPTY_DATA=/algod/empty

# To allow mounting the data directory we need to change permissions
# to our algorand user. The script is initially run as the root user
# in order to change permissions, afterwards the script is re-launched
# as the algorand user.
if [ "$(id -u)" = '0' ]; then
  chown -R algorand:algorand $EMPTY_DATA
  exec su -p -c "$(readlink -f $0) $@" algorand
fi


# Configure the participation node
if [ -d "$EMPTY_DATA" ]; then
  if [ "$TOKEN" != "" ]; then
      echo "$TOKEN" > "$EMPTY_DATA/algod.token"
  fi
  if [ "$ADMIN_TOKEN" != "" ]; then
    echo "$ADMIN_TOKEN" > "$EMPTY_DATA/algod.admin.token"
  fi
  cd $EMPTY_DATA
  cp "/node/run/genesis/testnet/genesis.json" genesis.json
  algocfg profile set --yes -d "$EMPTY_DATA" "participation"
  algod -o -d $EMPTY_DATA -l "0.0.0.0:8082"
else
  echo $EMPTY_DATA does not exist
  exit 1
fi