#!/usr/bin/env bash
# Stages the deterministic `statewalker stage demo` network for tape capture
# and records its data dir in .tapes/.datadir for the tape shell.
#
# Usage: .tapes/setup.sh   (normally invoked via `make tapes`)
#   TAPE_SPEED=fast|real   round-time profile (default: real, MainNet cadence)
set -euo pipefail

cd "$(dirname "$0")/.."

make statewalker build

DATADIR=$(./bin/statewalker stage demo --speed "${TAPE_SPEED:-real}" | awk -F= '/^DATADIR=/{print $2}')
if [ -z "$DATADIR" ]; then
    echo "stage demo did not print DATADIR; see the log above" >&2
    exit 1
fi

echo "$DATADIR" > .tapes/.datadir
echo "Demo network staged; DATADIR=$DATADIR"
