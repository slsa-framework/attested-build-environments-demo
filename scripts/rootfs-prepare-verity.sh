#!/bin/bash

set -e

echo Installing software necessary for verity measurement
sudo apt update
sudo apt-get install -y expect cryptsetup

echo "Fixing up GPT for the expanded disk"
expect -c 'spawn sudo parted /dev/disk/azure/scsi1/lun0 print; expect "Warning: Not all of the space available*"; send "f\r"; expect eof'

# sfdisk provides sector already aligned to 2048 (unlike parted)
START_POS=$(sudo sfdisk /dev/disk/azure/scsi1/lun0 -F | tail -n 1 | awk '{print $1}')
echo "Creating Verity device partition at $START_POS"
sudo parted -s /dev/disk/azure/scsi1/lun0 mkpart verity-device "${START_POS}s" 100%
sudo parted -s /dev/disk/azure/scsi1/lun0 print
