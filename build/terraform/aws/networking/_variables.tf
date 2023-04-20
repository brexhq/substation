variable "instance_tenancy" {
  description = "A tenancy option for instances launched into the VPC"
  type        = string
  default     = "default"
}


variable "private_subnet_cidr" {
  type    = string
  default = "10.0.0.0/17"
}

variable "public_subnet_cidr" {
  type    = string
  default = "10.0.128.0/17"
}

variable "availability_zone" {
  type    = string
  default = "us-east-1a"
}

variable "tags" {
  type    = map(any)
  default = {}
}
