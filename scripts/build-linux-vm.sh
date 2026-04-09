#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

VM_USER="${AZURE_VM_USER:-azureuser}"
SSH_KEYS_URL="${SSH_KEYS_URL:-https://github.com/$VM_USER.keys}"

echo Installing software desired for the eventual image
apt-get update
apt-get install -y golang tpm2-tools

echo "Setting public keys from $SSH_KEYS_URL"
mkdir -p /home/$VM_USER/.ssh
touch /home/$VM_USER/.ssh/authorized_keys
curl $SSH_KEYS_URL >> /home/$VM_USER/.ssh/authorized_keys
chown -R $VM_USER:$VM_USER /home/$VM_USER/.ssh
chmod 700 /home/$VM_USER/.ssh
chmod 600 /home/$VM_USER/.ssh/authorized_keys

echo Remove apt postinstall steps that impact the boot flow
rm /etc/kernel/postinst.d/zz-update-grub
rm /etc/kernel/postinst.d/initramfs-tools

echo Copying attestation utilities to sbin
chmod +x image-attestation
cp image-attestation /usr/sbin/image-attestation

echo Installing enlightened initramfs scripts and generate initramfs
TMP_DRIVE_PATH=$(mktemp -d)
"$SCRIPTPATH"/../initramfs/install.sh
mkinitramfs -o "$TMP_DRIVE_PATH/initrd.img-$(uname -r)"

echo Copying the kernel
cp "/boot/vmlinuz-$(uname -r)" $TMP_DRIVE_PATH

echo Creating tarball
tar -czf "$SCRIPTPATH"/image.tar.gz -C $TMP_DRIVE_PATH .

echo Updating initramfs
sudo cp "$TMP_DRIVE_PATH/initrd.img-$(uname -r)" /boot/

echo Enabling initramfs
sudo sed -i '/^GRUB_FORCE_PARTUUID/ s/^/#/' /etc/default/grub.d/40-force-partuuid.cfg
#sudo sed -i 's/^GRUB_RECORDFAIL_TIMEOUT=.*/GRUB_RECORDFAIL_TIMEOUT=0/' /etc/default/grub.d/50-cloudimg-settings.cfg
sudo update-grub

echo Disabling grubenv
sudo rm /boot/grub/grubenv
