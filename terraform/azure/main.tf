terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = var.name
  location = var.region
}

# Network Security Group
resource "azurerm_network_security_group" "main" {
  name                = var.name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}


resource "azurerm_network_security_rule" "aapptoolkit_api" {
  name                        = "AapptoolkitAPI"
  priority                    = 200
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range          = "*"
  destination_port_range     = "34000-34003"
  source_address_prefix      = "*"
  destination_address_prefix = "*"
  resource_group_name        = azurerm_resource_group.main.name
  network_security_group_name = azurerm_network_security_group.main.name
}

# Additional ports if specified
resource "azurerm_network_security_rule" "additional_ports" {
  count                       = length(var.additional_ports)
  name                        = "Port_${var.additional_ports[count.index]}"
  priority                    = 300 + count.index
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range          = "*"
  destination_port_range     = var.additional_ports[count.index]
  source_address_prefix      = "*"
  destination_address_prefix = "*"
  resource_group_name        = azurerm_resource_group.main.name
  network_security_group_name = azurerm_network_security_group.main.name
}

# Virtual Network
resource "azurerm_virtual_network" "main" {
  name                = var.name
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_subnet" "main" {
  name                 = "${var.name}-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.1.0/24"]
}

# VM
resource "azurerm_virtual_machine" "main" {
  name                  = var.name
  location              = azurerm_resource_group.main.location
  resource_group_name   = azurerm_resource_group.main.name
  network_interface_ids = [azurerm_network_interface.main.id]
  vm_size              = var.machine_type

  storage_os_disk {
    name              = var.name
    caching           = "ReadWrite"
    create_option     = "Attach"
    managed_disk_id   = azurerm_managed_disk.main.id
    disk_size_gb      = var.disk_size_gb
    os_type           = "Linux"
  }

  os_profile {
    computer_name  = var.name
    admin_username = "adminuser"
    custom_data    = var.aapp_manifest_yaml
  }

  os_profile_linux_config {
    disable_password_authentication = true
  }

  security_profile {
    security_type = "ConfidentialVM"
    vtpm_enabled  = true
    secure_boot_enabled = false
  }
}

resource "azurerm_network_interface" "main" {
  name                = var.name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.main.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.main.id
  }
}

resource "azurerm_public_ip" "main" {
  name                = var.name
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  allocation_method   = "Dynamic"
}

resource "azurerm_managed_disk" "main" {
  name                 = var.name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  storage_account_type = "Standard_LRS"
  create_option       = "Import"
  source_uri          = var.image_path
  os_type             = "Linux"
  disk_size_gb        = var.disk_size_gb
  security_type       = "ConfidentialVM_NonPersistedTPM"
  hyper_v_generation  = "V2"
}
