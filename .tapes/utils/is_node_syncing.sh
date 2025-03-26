#!/usr/bin/env bash

LOGPFX=$(basename "$0"):
cd "$(dirname "$(realpath "$0")")/.."

set -e

source utils/_common.sh

./utils/is_node_running.sh

# Sync time can glitch to 0.0 momentarily; query twice

if ! $GOAL_CMD node status | grep -q 'Sync Time: 0.0s'; then
    exit 0
fi

if ! goal node wait -w 6 > /dev/null 2>&1 ; then
    exit 0
fi

! $GOAL_CMD node status | grep -q 'Sync Time: 0.0s'