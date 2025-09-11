#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"

if ! [ -f ~/.ssh/id_rsa ]; then
    ssh-keygen -t rsa -b 2048 -f ~/.ssh/id_rsa -N ""
fi

az account set --subscription $AZURE_SUBSCRIPTION_ID
az group create --resource-group $AZURE_RESOURCE_GROUP --location $AZURE_LOCATION

echo "Creating gallery..."
az sig create --resource-group $AZURE_RESOURCE_GROUP \
              --gallery-name $AZURE_GALLERY_NAME

echo "Creating image definition..."
az sig image-definition create --resource-group $AZURE_RESOURCE_GROUP \
                               --gallery-name $AZURE_GALLERY_NAME \
                               --gallery-image-definition $AZURE_IMAGE_DEFINITION \
                               --os-state Specialized \
                               --os-type Linux \
                               --features SecurityType=TrustedLaunch \
                               --publisher DemoPublisher \
                               --offer DemoOffer \
                               --sku DemoSku

echo "Creating VM..."
az vm create --resource-group $AZURE_RESOURCE_GROUP \
             --name $AZURE_VM_NAME \
             --image Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest \
             --size Standard_D4ds_v5 \
             --admin-username azureuser \
             --ssh-key-value ~/.ssh/id_rsa.pub \
             --security-type TrustedLaunch \
             --nic-delete-option delete \
             --os-disk-delete-option delete \
             --patch-mode ImageDefault \
             | tee create.log

cleanup() {
    echo "Cleaning up..."
    #az vm delete --resource-group $AZURE_RESOURCE_GROUP --name $AZURE_VM_NAME --yes
    rm -f create.log
}

trap 'cleanup' EXIT

IP_ADDR=$(cat create.log | jq -r .publicIpAddress | tail -n 1)
echo "VM created with IP address: $IP_ADDR"

echo "Making sure we can connect to the VM..."
MAX_RETRIES=10
RETRY_DELAY=10
count=0
while [ $count -lt $MAX_RETRIES ]; do
    ssh -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa azureuser@$IP_ADDR "uname -a" && break
    count=$((count + 1))
    echo "Retry $count/$MAX_RETRIES failed. Waiting $RETRY_DELAY seconds before next attempt..."
    sleep $RETRY_DELAY
done

if [ $count -eq $MAX_RETRIES ]; then
    echo "All $MAX_RETRIES attempts failed."
    exit 1
fi

echo "Copying files to VM..."
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../initramfs" azureuser@$IP_ADDR:
scp -r -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/../scripts"  azureuser@$IP_ADDR:
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa "$SCRIPTPATH/image-attestation"  azureuser@$IP_ADDR:

echo "Building VM image..."
ssh    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa azureuser@$IP_ADDR "sudo scripts/build-linux-vm.sh"
scp    -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa azureuser@$IP_ADDR:~/scripts/image.tar.gz .

echo "Creating image version..."
VM_ID=$(az vm show --name $AZURE_VM_NAME --resource-group $AZURE_RESOURCE_GROUP --query id -o tsv)
az sig image-version create --resource-group $AZURE_RESOURCE_GROUP \
                            --gallery-name $AZURE_GALLERY_NAME \
                            --gallery-image-definition $AZURE_IMAGE_DEFINITION \
                            --gallery-image-version $AZURE_IMAGE_VERSION \
                            --virtual-machine $VM_ID
