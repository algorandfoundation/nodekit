#!/usr/bin/env bash

# Base data directory, uses ALGORAND_DATA if set, otherwise default to /var/lib/..
export DATA_DIR=${ALGORAND_DATA:-/var/lib/algorand}

# network data dir under $DATA_DIR
export NETWORK=fnet-v1

# DB dir to nuke when reseting
export DB_DIR="$DATA_DIR/$NETWORK"

# User owner of files. Should resolve automatically
export USER=$(stat -c '%U' "$DATA_DIR/genesis.json")

# Group owner of files. Should resolve automatically
export USERGRP=$(stat -c '%G' "$DATA_DIR/genesis.json")

export GOAL_CMD="goal -d $DATA_DIR"

export ALGORAND_DATA="$DATA_DIR"

EXPECTED_ARCH="x86_64"

if [ "$(uname -i)" != "$EXPECTED_ARCH" ]; then
    echo "Unexpected architecture, wanted $EXPECTED_ARCH, found $(uname -i)"
    echo "These scripts only work for x86_64"
    exit 1
fi

function confirm_requirements {
  # check that requirements are installed
  # override default list by calling with arguments
  reqs=${@:-curl dig sha256sum md5sum jq tr cut sed shuf wc}
  # echo -n "confirm_requirements: " >&2
  for req in $reqs; do
    if ! which "$req" > /dev/null 2>&1; then
      echo "Error: '$req' is required but not installed" >&2
      exit 1
    fi
  done
}

confirm_requirements

function md5 {
    md5sum "$1" | cut -d\  -f1
}
export -f md5

function sha256 {
    sha256sum "$1" | cut -d\  -f1
}
export -f sha256