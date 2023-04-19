variable "vpc_network_cidr" {
  description = "IP CIDR for the vpc"
  type = string
  default = "10.0.0.0/16"
}

variable "instance_tenancy" {
  description = "A tenancy option for instances launched into the VPC"
  type        = string
  default     = "default"
}

variable "enable_flow_log" {
  description = "Whether or not to enable VPC Flow Logs"
  type        = bool
  default     = false
}

variable "tags" {
  type    = map(any)
  default = {}
}