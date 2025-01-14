#!/usr/bin/env bash

set -euo pipefail
  
os=$(uname -ms)
release="https://github.com/algorandfoundation/nodekit/releases/latest/download"

Red=''
Green=''
Yellow=''
Blue=''
Opaque=''
Bold_White=''
Bold_Green=''
Reset=''

if [[ -t 1 ]]; then
  Reset='\033[0m'
  Red='\033[0;31m'
  Green='\033[0;32m'
  Yellow='\033[0;33m'
  Blue='\033[0;34m'
  Bold_Green='\033[1;32m'
  Bold_White='\033[1m'
  Opaque='\033[0;2m'
fi

success() {
  echo -e "${Green}$@ ${Reset}"
}

info() {
  echo -e "${Opaque}$@ ${Reset}"
}

warn() {
  echo -e "${Yellow}WARN${Reset}: ${Opaque}$@ ${Reset}"
}

error() {
  echo -e "${Red}ERROR${Reset}:" "${Yellow}" "$@" "${Reset}" >&2
  exit 1
}

prompt_default_no() {
  echo -n "$1 (y/N) "
  read -rn1 prompt_response < /dev/tty
  echo ""
  if [[ "$prompt_response" =~ "y" ]]; then
	  return 0
  fi
  return 1
}

if [ -f nodekit ]; then
  warn "A nodekit file already exists in the current directory."
  if prompt_default_no "Do you want to upgrade it to the latest nodekit?"; then
    rm nodekit
  else
    info "Not upgrading nodekit.\n\nYou can run nodekit with:\n\n./nodekit\n\nOr start the installer with:\n\n./nodekit bootstrap"
    exit 0
  fi
fi

if [[ ${OS:-} = Windows_NT ]]; then
  error "Unsupported platform. Windows is supported only via WSL.\n\nFor WSL installation instructions, see here: https://learn.microsoft.com/en-us/windows/wsl/install"
fi

trap "warn SIGINT received." int
trap "info Exiting the installation" exit

case $os in
'Darwin x86_64')
  platform=amd64-darwin
  ;;
'Darwin arm64')
  platform=arm64-darwin
  ;;
'Linux aarch64' | 'Linux arm64')
  platform=arm64-linux
  ;;
'Linux x86_64' | *)
  platform=amd64-linux
  ;;
esac
 
target="nodekit-$platform"
url="$release/$target"

echo -e "${Opaque}Downloading:${Reset}${Bold_White} $target ${Reset}from $url"
curl --fail --location --progress-bar --output nodekit "$url" ||
error "Failed to download ${target} from ${release} ${url}"

chmod +x nodekit

trap - int
trap - exit

success "Downloaded: ${Bold_Green}${target} as nodekit 🎉${Reset}"
info "Explore all nodekit options with:"
echo "./nodekit --help"
echo ""
info "Starting nodekit bootstrap"
echo "./nodekit bootstrap"

./nodekit bootstrap
