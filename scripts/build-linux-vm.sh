#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

VM_USER="${VM_USER:-azureuser}"
SSH_KEYS_URL="${SSH_KEYS_URL:-https://github.com/marcelamelara.keys}"

echo Installing software necessary for image build
apt-get update
apt-get install -y rsync

echo Installing software desired for the eventual image
apt-get install -y golang tpm2-tools

echo Setting public keys
mkdir -p /home/$VM_USER/.ssh
touch /home/$VM_USER/.ssh/authorized_keys
curl $SSH_KEYS_URL >> /home/$VM_USER/.ssh/authorized_keys
chown -R $VM_USER:$VM_USER /home/$VM_USER/.ssh

echo Remove apt postinstall steps that impact the boot flow
rm /etc/kernel/postinst.d/zz-update-grub
rm /etc/kernel/postinst.d/initramfs-tools

echo Copying attestation utilities to sbin
chmod +x image-attestation
cp image-attestation /usr/sbin/image-attestation

echo Measuring rootfs
"$SCRIPTPATH"/scripts/rootfs-measure-verity.sh

echo Installing enlightened initramfs scripts and generate initramfs
"$SCRIPTPATH"/initramfs/install.sh
mkinitramfs -o "$TMP_DRIVE_PATH/initrd-$(uname -r).img"

echo Copying the kernel
cp "/boot/vmlinuz-$(uname -r)" $TMP_DRIVE_PATH

echo Creating tarball
tar -czf "$SCRIPTPATH"/image.tar.gz -C $TMP_DRIVE_PATH .
