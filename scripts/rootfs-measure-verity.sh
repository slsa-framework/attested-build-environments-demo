#!/bin/bash

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

VM_USER="${VM_USER:-azureuser}"

echo Installing software necessary for verity measurement
apt-get update
apt-get install -y cryptsetup

echo Patching up fstab
# Use UEFI label for the EFI partition instead of UUID
sed -i 's/UUID=[^\s]\+\(\s\+\/boot\/efi\)/LABEL=UEFI\1/' /etc/fstab
# Remove the /mnt partition, it should already be used for overlay
sed -i '/\/mnt/d' /etc/fstab

echo Create and mount ext4 volume
TMP_DRIVE_PATH="/mnt/fs-tmp"
FS_FILE="$TMP_DRIVE_PATH/fs.img"
FS_MOUNT="$TMP_DRIVE_PATH/fs"
FS_SIZE="3G"

mkdir -p $TMP_DRIVE_PATH
rm -rf $TMP_DRIVE_PATH/*
truncate -s $FS_SIZE $FS_FILE
mkfs.ext4 $FS_FILE
tune2fs -c 0 -i 0 $FS_FILE
mkdir -p $FS_MOUNT
mount $FS_FILE $FS_MOUNT

# copy full rootfs into new fs volume (using -x to avoid cross fs boundaries)
# rsync over cp since cp has issues handling copies from /
# you might normally use /* to avoid this, but that starts pulling from /proc and
# other special filesystems that we don't want to copy
echo Copying rootfs into new volume
rsync -ax / $FS_MOUNT/

echo Cleaning up file from new volume
rm -rf $FS_MOUNT/tmp/*
rm -rf $FS_MOUNT/home/$VM_USER/* # this glob should leave the .ssh directory

echo "Add a marker file to show that we're in an attested VM"
touch $FS_MOUNT/home/$VM_USER/attested-vm

echo Unmounting fs volume
umount $FS_MOUNT
rm -r $FS_MOUNT

echo Generating verity files
# Verity root hash will be in /measurements/eventlog if using enlightened initramfs
veritysetup format $FS_FILE $TMP_DRIVE_PATH/fs-verity.img --root-hash-file $TMP_DRIVE_PATH/fs.hash
