#!/bin/bash

set -e

# TODO: pass it into this script as parameter
LUN_ID="0"
# TODO: Look it up by 'cloudimg-rootfs' partition label
ROOTFS_DEVICE="/dev/disk/azure/scsi1/lun${LUN_ID}-part1"
# TODO: Look it up by 'verity-device' GPT partition name
VERITY_DEVICE="/dev/disk/azure/scsi1/lun${LUN_ID}-part2"
# TODO: Look it up by 'UEFI' partition label
UEFI_DEVICE="/dev/disk/azure/scsi1/lun${LUN_ID}-part15"

sudo mkdir -p /mnt/uefi
sudo mount $UEFI_DEVICE /mnt/uefi

echo "Setting up Verity for $ROOTFS_DEVICE on $VERITY_DEVICE"
sudo mkdir -p /mnt/uefi/verity
sudo veritysetup --verbose --debug format /dev/disk/azure/scsi1/lun0-part1 /dev/disk/azure/scsi1/lun0-part2 --root-hash-file /mnt/uefi/verity/rootfs.hash

blkid -s UUID -o value $ROOTFS_DEVICE | sudo tee /mnt/uefi/verity/rootfs.uuid
blkid -s UUID -o value $VERITY_DEVICE | sudo tee /mnt/uefi/verity/verityfs.uuid
echo "slsa-verity" | sudo tee /mnt/uefi/verity/verity.name

sudo umount /mnt/uefi
