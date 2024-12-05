#!/usr/bin/env bash

set -euo pipefail

if [ -f algorun ]; then
    echo An algorun file already exists in the current directory. Delete or rename it before installing.
    exit 1
fi

os=$(uname -ms)

release="https://github.com/algorandfoundation/hack-tui/releases/download"
version="v1.0.0-beta.1"

if [[ ${OS:-} = Windows_NT ]]; then
  echo "Unsupported platform"
  exit 1
fi

trap "echo Something went wrong." int
trap "echo Something went wrong." exit

case $os in
'Darwin x86_64')
    target=algorun-amd64-darwin
    ;;
'Darwin arm64')
    target=algorun-arm64-darwin
    ;;
'Linux aarch64' | 'Linux arm64')
    target=algorun-arm64-linux
    ;;
'Linux x86_64' | *)
    target=algorun-amd64-linux
    ;;
esac

echo "Downloading: $release/$version/$target"
curl --fail --location --progress-bar --output algorun "$release/$version/$target"

chmod +x algorun

trap - int
trap - exit

echo "Downloaded"
echo "Run with:"
echo "./algorun"
