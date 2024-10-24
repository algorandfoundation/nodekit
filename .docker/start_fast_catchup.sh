#!/usr/bin/env bash

set -e

if [ "$DEBUG" = "1" ]; then
  set -x
fi

if [ "$ALGORAND_DATA" != "/algod/data" ]; then
  echo "Do not override 'ALGORAND_DATA' environment variable."
  exit 1
fi

FAST_CATCHUP_DATA=/algod/fast-catchup

# To allow mounting the data directory we need to change permissions
# to our algorand user. The script is initially run as the root user
# in order to change permissions, afterwards the script is re-launched
# as the algorand user.
if [ "$(id -u)" = '0' ]; then
  chown -R algorand:algorand $FAST_CATCHUP_DATA
  exec su -p -c "$(readlink -f $0) $@" algorand
fi

function catchup() {
  sleep 5
  goal node catchup --force --min 1000000 -d $FAST_CATCHUP_DATA
}

# Configure the participation node
if [ -d "$FAST_CATCHUP_DATA" ]; then
  if [ "$TOKEN" != "" ]; then
      echo "$TOKEN" > "$FAST_CATCHUP_DATA/algod.token"
  fi
  if [ "$ADMIN_TOKEN" != "" ]; then
    echo "$ADMIN_TOKEN" > "$FAST_CATCHUP_DATA/algod.admin.token"
  fi
  cd $FAST_CATCHUP_DATA
  cp "/node/run/genesis/testnet/genesis.json" genesis.json
  algocfg profile set --yes -d "$FAST_CATCHUP_DATA" "participation"
  catchup &
  algod -o -d $FAST_CATCHUP_DATA -l "0.0.0.0:8081"
else
  echo $FAST_CATCHUP_DATA does not exist
  exit 1
fi