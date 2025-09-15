#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"
VM_USER="${AZURE_USER_NAME:-azureuser}"

echo "Creating image VM..."
IMAGE_VM_NAME="${AZURE_VM_NAME}image"
$SCRIPTPATH/create-vm $IMAGE_VM_NAME
IMAGE_VM_ID=$(az vm show --resource-group $AZURE_RESOURCE_GROUP --name $IMAGE_VM_NAME | jq -r ".id")
IP_ADDR=$($SCRIPTPATH/get-ip $IMAGE_VM_ID)
$SCRIPTPATH/test-connectivity $IP_ADDR

echo "Copying files to VM..."
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../initramfs" "${VM_USER}@${IP_ADDR}":
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../scripts"  "${VM_USER}@${IP_ADDR}":
#scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/image-attestation"  "${VM_USER}@${IP_ADDR}":

#echo "Building VM image..."
ssh    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "${VM_USER}@${IP_ADDR}" "sudo scripts/build-linux-vm.sh"
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "${VM_USER}@${IP_ADDR}":~/scripts/image.tar.gz .

echo "Deleting image VM and detaching OS disk"
DISK_ID=$(az vm show --id $IMAGE_VM_ID | jq -r ".storageProfile.osDisk.managedDisk.id")
az vm delete --id $IMAGE_VM_ID --yes

echo "Creating hasher VM..."
HASHER_VM_NAME="${AZURE_VM_NAME}hash"
$SCRIPTPATH/create-vm $HASHER_VM_NAME
HASHER_VM_ID=$(az vm show --resource-group $AZURE_RESOURCE_GROUP --name $HASHER_VM_NAME | jq -r ".id")
IP_ADDR=$($SCRIPTPATH/get-ip $HASHER_VM_ID)
$SCRIPTPATH/test-connectivity $IP_ADDR

echo "Attaching image disk (with 10% added space for verity hashes)"
DISK_SIZE=$(az disk show --id $DISK_ID --query "diskSizeGB")
NEW_DISK_SIZE=$(echo "$DISK_SIZE * 1.1" | bc)
NEW_DISK_SIZE=$(printf "%.0f" "$NEW_DISK_SIZE")
az disk update --id $DISK_ID --disk-size-gb $NEW_DISK_SIZE
az vm disk attach --resource-group $AZURE_RESOURCE_GROUP --vm-name $HASHER_VM_NAME --disk-id $DISK_ID --lun 0

echo "Setting up Verity..."
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../scripts"  "${VM_USER}@${IP_ADDR}":
ssh    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "${VM_USER}@${IP_ADDR}" "sudo scripts/rootfs-prepare-verity.sh"
ssh    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "${VM_USER}@${IP_ADDR}" "sudo scripts/rootfs-measure-verity.sh"

echo "Deallocating hasher VM"
az vm deallocate --id $HASHER_VM_ID
