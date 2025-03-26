#!/usr/bin/env bash

LOGPFX=$(basename "$0"):
cd "$(dirname "$(realpath "$0")")/.."

set -e

source utils/_common.sh

PID="$DATA_DIR/algod.pid"

if [ ! -f "$PID" ]; then
    exit 1
fi

# Is algod booted inside container?
$GOAL_CMD node status > /dev/null 2>&1

# Give a second or two for Sync time to budge from zero
# if we are not synced
sleep 1