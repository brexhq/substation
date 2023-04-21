variable "cidr_block" {
  description = "The CIDR block for the VPC."
  type        = string
  # 16k hosts per subnet
  default = "10.0.0.0/16"
}

variable "public_subnet" {
  description = "The public subnet to use for the VPC."
  type        = map(string)
  # 16k hosts per subnet
  default = {
    "10.0.0.0/18" = "us-east-1a"
  }

  validation {
    condition     = length(keys(var.public_subnet)) == 1
    error_message = "Only one public subnet is allowed."
  }
}

variable "private_subnets" {
  description = "The private subnets to use for the VPC."
  type        = map(string)
  # 16k hosts per subnet
  default = {
    "10.0.64.0/18"  = "us-east-1a"
    "10.0.128.0/18" = "us-east-1b"
    "10.0.192.0/18" = "us-east-1c"
  }

  validation {
    condition     = length(keys(var.private_subnets)) >= 1
    error_message = "At least one private subnet is required."
  }
}

variable "tags" {
  description = "Tags to apply to the VPC."
  type        = map(any)
  default     = {}
}
