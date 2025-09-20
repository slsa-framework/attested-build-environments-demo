#!/bin/bash

set -e

# TODO: pass it into this script as parameter
LUN_ID="0"
# TODO: Look it up by 'cloudimg-rootfs' partition label
ROOTFS_DEVICE="/dev/disk/azure/scsi1/lun${LUN_ID}-part1"
# TODO: Look it up by 'verity-device' GPT partition name
VERITY_DEVICE="/dev/disk/azure/scsi1/lun${LUN_ID}-part2"

echo "Setting up Verity for $ROOTFS_DEVICE on $VERITY_DEVICE"
sudo mkdir -p /boot/efi/verity
sudo veritysetup --verbose --debug format /dev/disk/azure/scsi1/lun0-part1 /dev/disk/azure/scsi1/lun0-part2 --root-hash-file /boot/efi/verity/rootfs.hash

blkid -s PARTUUID -o value $ROOTFS_DEVICE > /boot/efi/verity/rootfs-device.uuid
blkid -s PARTUUID -o value $VERITY_DEVICE > /boot/efi/verity/verity-device.uuid
