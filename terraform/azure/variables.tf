variable "name" {
  description = "The name prefix for all resources"
  type        = string
}

variable "region" {
  description = "The Azure region to deploy to"
  type        = string
  default     = "westeurope"
}

variable "machine_type" {
  description = "The VM size to use"
  type        = string
  default     = "Standard_EC4eds_v5"
}


variable "additional_ports" {
  description = "List of additional ports to open"
  type        = list(string)
  default     = []
}

variable "image_path" {
  description = "The path to the VM image"
  type        = string
}

variable "disk_size_gb" {
  description = "The size of the OS disk in GB"
  type        = number
  default     = 30
}


variable "aapp_manifest_yaml" {
  description = "The aApp manifest YAML created by sops"
  type        = string
  default     = ""
}
