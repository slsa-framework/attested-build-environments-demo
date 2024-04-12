#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

echo Installing software necessary for image build
apt-get update
apt-get install -y rsync cryptsetup

echo Installing software desired for the eventual image
apt-get install -y golang tpm2-tools

echo Setting public keys
mkdir -p /home/azureuser/.ssh
touch /home/azureuser/.ssh/authorized_keys
curl https://github.com/chkimes.keys >> /home/azureuser/.ssh/authorized_keys
curl https://github.com/marcelamelara.keys >> /home/azureuser/.ssh/authorized_keys
chown -R azureuser:azureuser /home/azureuser/.ssh

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

echo Unmounting fs volume
umount $FS_MOUNT
rm -r $FS_MOUNT

echo Generating verity files
veritysetup format $FS_FILE $TMP_DRIVE_PATH/fs-verity.img --root-hash-file $TMP_DRIVE_PATH/fs.hash

echo Installing initramfs scripts and generate initramfs
$SCRIPTPATH/initramfs/install.sh
mkinitramfs -o $TMP_DRIVE_PATH/initrd-$(uname -r).img

echo Copying the kernel
cp /boot/vmlinuz-$(uname -r) $TMP_DRIVE_PATH

echo Creating tarball
tar -czf $SCRIPTPATH/image.tar.gz -C $TMP_DRIVE_PATH .
