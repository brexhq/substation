variable "instance_tenancy" {
  description = "A tenancy option for instances launched into the VPC"
  type        = string
  default     = "default"
}

variable "tags" {
  type    = map(any)
  default = {}
}

variable "subnet_cidr" {
  type    = string
  default = "10.0.0.0/16"
}