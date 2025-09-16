#!/bin/bash

set -e

SCRIPTPATH="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P )"

echo "Creating attested VM..."
ATTEST_VM_NAME="${AZURE_VM_NAME}attest"
IMAGE_RESOURCE=$($SCRIPTPATH/get-image)
$SCRIPTPATH/create-vm $ATTEST_VM_NAME $IMAGE_RESOURCE
ATTEST_VM_ID=$(az vm show --resource-group $AZURE_RESOURCE_GROUP --name $ATTEST_VM_NAME | jq -r ".id")

echo "Attesting VM..."
#TODO

echo "Deallocating attested VM"
az vm deallocate --id $ATTEST_VM_ID
