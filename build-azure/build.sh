#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"

if ! [ -f ~/.ssh/id_rsa ]; then
    ssh-keygen -t rsa -b 2048 -f ~/.ssh/id_rsa -N ""
fi

az account set --subscription $AZURE_SUBSCRIPTION_ID
az group create --resource-group $AZURE_RESOURCE_GROUP --location $AZURE_LOCATION

echo "Creating VM..."
az vm create --resource-group $AZURE_RESOURCE_GROUP \
             --name $AZURE_VM_NAME \
             --image Canonical:0001-com-ubuntu-minimal-mantic:minimal-23_10-gen2:23.10.202402260 \
             --size Standard_D4ds_v5 \
             --admin-username azureuser \
             --ssh-key-value ~/.ssh/id_rsa.pub \
             --security-type TrustedLaunch \
             --nic-delete-option delete \
             --os-disk-delete-option delete \
             | tee create.log

cleanup() {
    echo "Cleaning up..."
    az vm delete --resource-group $AZURE_RESOURCE_GROUP --name $AZURE_VM_NAME --yes
    rm -f create.log
}

trap 'cleanup' EXIT

IP_ADDR=$(cat create.log | jq -r .publicIpAddress | tail -n 1)
echo "VM created with IP address: $IP_ADDR"

echo "Copying files to VM..."
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../initramfs" azureuser@$IP_ADDR:
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/build-vm.sh"  azureuser@$IP_ADDR:
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/client"  azureuser@$IP_ADDR:
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/server"  azureuser@$IP_ADDR:

echo "Building VM image..."
ssh    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa azureuser@$IP_ADDR "sudo ./build-vm.sh"
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa azureuser@$IP_ADDR:~/image.tar.gz .
