variable "config" {
  type = object({
    name   = string
    policy = optional(string, null)
  })
  description = <<EOH
    Configuration for the KMS key:

    * name: The name of the KMS key.
    * policy: The policy to attach to the KMS key. If not provided, then the default policy will be used.
EOH
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
