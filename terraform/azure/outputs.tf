output "public_ip" {
  description = "The public IP address of the VM"
  value       = azurerm_public_ip.main.ip_address
}

output "resource_group_name" {
  description = "The name of the resource group"
  value       = azurerm_resource_group.main.name
}
