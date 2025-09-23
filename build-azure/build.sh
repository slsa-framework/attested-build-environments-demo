#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"
VM_USER="${AZURE_VM_USER:-azureuser}"

echo "Creating resource group..."
az account set --subscription $AZURE_SUBSCRIPTION_ID
az group create --resource-group $AZURE_RESOURCE_GROUP --location $AZURE_LOCATION

#echo "Creating image gallery..."
#$SCRIPTPATH/create-gallery

echo "Creating image VM..."
IMAGE_VM_NAME="${AZURE_VM_NAME}image"
$SCRIPTPATH/create-vm $IMAGE_VM_NAME
IMAGE_VM_ID=$(az vm show --resource-group $AZURE_RESOURCE_GROUP --name $IMAGE_VM_NAME | jq -r ".id")
IP_ADDR=$($SCRIPTPATH/get-ip $IMAGE_VM_ID)
$SCRIPTPATH/test-connectivity $IP_ADDR

echo "Copying files to VM..."
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../initramfs" "${VM_USER}@${IP_ADDR}":
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../scripts"  "${VM_USER}@${IP_ADDR}":
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/image-attestation"  "${VM_USER}@${IP_ADDR}":

echo "Building VM image..."
ssh    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "${VM_USER}@${IP_ADDR}" "sudo SSH_KEYS_URL=$SSH_KEYS_URL scripts/build-linux-vm.sh"
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "${VM_USER}@${IP_ADDR}":~/scripts/image.tar.gz .

echo "Deallocating image VM..."
az vm deallocate --id $IMAGE_VM_ID

echo "Detachinging OS disk..."
DISK_ID=$(az vm show --id $IMAGE_VM_ID | jq -r ".storageProfile.osDisk.managedDisk.id")
IMAGE_ID=$(az disk show --id $DISK_ID | jq -r ".creationData.imageReference.id")
SWAP_DISK_NAME="${AZURE_VM_NAME}-$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9' | head -c 16 ; echo)"
az disk create --image-reference $IMAGE_ID --resource-group $AZURE_RESOURCE_GROUP --name $SWAP_DISK_NAME --security-type TrustedLaunch
SWAP_DISK_ID=$(az disk show --resource-group $AZURE_RESOURCE_GROUP --name $SWAP_DISK_NAME | jq -r ".id")
az vm update --name $IMAGE_VM_NAME --resource-group $AZURE_RESOURCE_GROUP --os-disk $SWAP_DISK_ID

echo "Creating hasher VM..."
HASHER_VM_NAME="${AZURE_VM_NAME}hash"
$SCRIPTPATH/create-vm $HASHER_VM_NAME
HASHER_VM_ID=$(az vm show --resource-group $AZURE_RESOURCE_GROUP --name $HASHER_VM_NAME | jq -r ".id")
IP_ADDR=$($SCRIPTPATH/get-ip $HASHER_VM_ID)
$SCRIPTPATH/test-connectivity $IP_ADDR

echo "Attaching image disk (with 10% added space for verity hashes)..."
DISK_SIZE=$(az disk show --id $DISK_ID --query "diskSizeGB")
NEW_DISK_SIZE=$(echo "$DISK_SIZE * 1.1" | bc)
NEW_DISK_SIZE=$(printf "%.0f" "$NEW_DISK_SIZE")
az disk update --id $DISK_ID --disk-size-gb $NEW_DISK_SIZE
az vm disk attach --resource-group $AZURE_RESOURCE_GROUP --vm-name $HASHER_VM_NAME --disk-id $DISK_ID --lun 0

echo "Setting up Verity..."
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../scripts"  "${VM_USER}@${IP_ADDR}":
ssh    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "${VM_USER}@${IP_ADDR}" "sudo scripts/rootfs-prepare-verity.sh"
ssh    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "${VM_USER}@${IP_ADDR}" "sudo scripts/rootfs-measure-verity.sh"

echo "Deleting hasher VM..."
az vm delete --id $HASHER_VM_ID --yes

#echo "Creating image version ...
#$SCRIPTPATH/create-image $DISK_ID

#echo "Creating disk copy..."
#CLONE_DISK_NAME="${AZURE_VM_NAME}-$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9' | head -c 16 ; echo)"
#az disk create --resource-group $AZURE_RESOURCE_GROUP --name $CLONE_DISK_NAME --source $DISK_ID
#az vm disk attach --resource-group $AZURE_RESOURCE_GROUP --vm-name $IMAGE_VM_NAME --name $CLONE_DISK_NAME --lun 0

echo "Attaching OS disk back..."
az vm update --name $IMAGE_VM_NAME --resource-group $AZURE_RESOURCE_GROUP --os-disk $DISK_ID
az disk delete --id $SWAP_DISK_ID --yes
