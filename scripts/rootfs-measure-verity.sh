#!/bin/bash

set -e

# Expected to have 'cloudimg-rootfs' partition label
ROOTFS_DEVICE="/dev/disk/azure/scsi1/lun0-part1"
# Expected to have 'verity-device' GPT partition name
VERITY_DEVICE="/dev/disk/azure/scsi1/lun0-part2"
echo "Setting up Verity for $ROOTFS_DEVICE on $VERITY_DEVICE"
sudo mkdir -p /boot/verity
sudo veritysetup --verbose --debug format /dev/disk/azure/scsi1/lun0-part1 /dev/disk/azure/scsi1/lun0-part2 --root-hash-file /boot/verity/fs.hash
