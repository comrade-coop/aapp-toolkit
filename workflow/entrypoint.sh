#!/bin/bash

# Exit on any error
set -e

# Access input parameters
CLIENT_ID="$1"
CLIENT_SECRET="$2"
SUBSCRIPTION_ID="$3"
TENANT_ID="$4"
RESOURCE_GROUP="$5"
VM_NAME="$6"
VM_NETWORK_ID="$7"
VM_DEV_KEY="$8"
AAPP_MANIFEST="$9"
WORKFLOW_PATH="${10}"

# Validate input parameters
if [ -z "$CLIENT_ID" ] || [ -z "$CLIENT_SECRET" ] || [ -z "$SUBSCRIPTION_ID" ] || [ -z "$TENANT_ID" ] || [ -z "$RESOURCE_GROUP" ] || [ -z "$VM_NAME" ] || [ -z "$VM_NETWORK_ID" ] || [ -z "$VM_DEV_KEY" ] || [ -z "$AAPP_MANIFEST" ] || [ -z "$WORKFLOW_PATH" ]; then
    echo "Error: All parameters must be provided"
    exit 1
fi

echo "Installing Azure CLI..."
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

echo "Logging in to Azure..."
az login --service-principal \
    --username "$CLIENT_ID" \
    --password "$CLIENT_SECRET" \
    --tenant "$TENANT_ID"

echo "Setting subscription..."
az account set --subscription "$SUBSCRIPTION_ID"

if [ ! -f "$AAPP_MANIFEST" ]; then
    echo "Error: Manifest file not found at $AAPP_MANIFEST"
    exit 1
fi

CLOUD_INIT="$WORKFLOW_PATH/../image/cloud-init.yml"
ARM_TEMPLATE="$WORKFLOW_PATH/template.json"

echo "Creating VM: $VM_NAME with cloud-init and app manifest"
az deployment group create \
    --resource-group $RESOURCE_GROUP \
    --template-file $ARM_TEMPLATE \
    --parameters \
                virtualMachineName="$VM_NAME" \
                virtualNetworkId="$VM_NETWORK_ID" \
                subnetName="default" \
                sshKeyName="$VM_DEV_KEY" \
                userData=@$AAPP_MANIFEST \
                customData=@$CLOUD_INIT

echo "VM creation completed successfully"