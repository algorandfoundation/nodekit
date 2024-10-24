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
  chown -R algorand:algorand $ALGORAND_DATA
  exec su -p -c "$(readlink -f $0) $@" algorand
fi

/node/run/start_empty.sh &
/node/run/start_fast_catchup.sh &
/node/run/start_dev.sh