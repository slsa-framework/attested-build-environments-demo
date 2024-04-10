#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

apt-get update
apt-get install -y initramfs-tools zstd tpm2-tools
cp -r "$SCRIPTPATH/initramfs-tools" /usr/share
