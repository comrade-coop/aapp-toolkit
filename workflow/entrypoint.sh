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
AAPP_MANIFEST="$7"

# Validate input parameters
if [ -z "$CLIENT_ID" ] || [ -z "$CLIENT_SECRET" ] || [ -z "$SUBSCRIPTION_ID" ] || [ -z "$TENANT_ID" ] || [ -z "$RESOURCE_GROUP" ] || [ -z "$VM_NAME" ] || [ -z "$AAPP_MANIFEST" ]; then
    echo "Error: All parameters must be provided"
    exit 1
fi

# Install Azure CLI
echo "Installing Azure CLI..."
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Login to Azure
echo "Logging in to Azure..."
az login --service-principal \
    --username "$CLIENT_ID" \
    --password "$CLIENT_SECRET" \
    --tenant "$TENANT_ID"

# Set subscription
echo "Setting subscription..."
az account set --subscription "$SUBSCRIPTION_ID"

# Validate manifest file exists
if [ ! -f "$AAPP_MANIFEST" ]; then
    echo "Error: Manifest file not found at $AAPP_MANIFEST"
    exit 1
fi

# Create VM
echo "Creating VM: $VM_NAME with cloud-init and app manifest"
az vm create \
    --resource-group "$RESOURCE_GROUP" \
    --name "$VM_NAME" \
    --image Ubuntu2204 \
    --admin-username azureuser \
    --generate-ssh-keys \
    --custom-data "@${GITHUB_ACTION_PATH}/../../image/cloud-init.yml" \
    --user-data "$AAPP_MANIFEST"

echo "VM creation completed successfully"
