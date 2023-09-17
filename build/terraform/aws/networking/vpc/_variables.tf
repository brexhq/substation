variable "config" {
  type = object({
    cidr_block = optional(string, "10.0.0.0/16")
    public_subnet = optional(map(string), {
      "10.0.0.0/18" = "us-east-1a"
    })
    private_subnets = optional(map(string), {
      "10.0.64.0/18"  = "us-east-1a"
      "10.0.128.0/18" = "us-east-1b"
      "10.0.192.0/18" = "us-east-1c"
    })
  })
  description = "Configuration for the VPC."

  validation {
    condition     = length(keys(var.config.public_subnet)) == 1
    error_message = "Only one public subnet is allowed."
  }

  validation {
    condition     = length(keys(var.config.private_subnets)) >= 1
    error_message = "At least one private subnet is required."
  }
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
