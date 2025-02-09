# Azure VM Creation GitHub Action

This action provides a reusable workflow component that creates an Azure VM in a specified resource group.

## Prerequisites

- Azure subscription
- Azure App registration with appropriate permissions
- Resource group already created in Azure

## Inputs

| Input | Description | Required |
|-------|-------------|----------|
| client_id | Azure App Registration Client ID | Yes |
| client_secret | Azure App Registration Client Secret | Yes |
| subscription_id | Azure Subscription ID | Yes |
| resource_group | Azure Resource Group Name | Yes |
| vm_name | Name for the new VM | Yes |
| aapp_manifest | Path to YAML file containing VM initialization data | Yes |

## Usage

```yaml
steps:
  - uses: comrade-coop/aapp-toolkit/workflow@v1
    with:
      client_id: ${{ secrets.AZURE_CLIENT_ID }}
      client_secret: ${{ secrets.AZURE_CLIENT_SECRET }}
      subscription_id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
      resource_group: 'my-resource-group'
      vm_name: 'my-vm-name'
      aapp_manifest: './aapp_manifest.yml'
```

## Example Manifest YAML
```yaml
#cloud-config
packages:
  - docker.io
  - docker-compose
runcmd:
  - systemctl start docker
  - systemctl enable docker
```

## Example Workflow

```yaml
name: Create Azure VM
on: [workflow_dispatch]

jobs:
  create-vm:
    runs-on: ubuntu-latest
    steps:
      - uses: comrade-coop/aapp-toolkit/workflow@v1
        with:
          client_id: ${{ secrets.AZURE_CLIENT_ID }}
          client_secret: ${{ secrets.AZURE_CLIENT_SECRET }}
          subscription_id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
          resource_group: 'development-rg'
          vm_name: 'test-vm-001'
```

## Versioning

This action follows semantic versioning. Use the appropriate version tag to ensure stability in your workflows:

- Use `@v1.0.0` for a specific version
- Use `@v1` for the latest v1.x.x release
- Use `@main` for the latest code (not recommended for production)

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## Contributing

This action is part of the aApp Toolkit project. For contribution guidelines, please refer to the main [README](../README.md#🤝-contributing).
