variable "config" {
  type = object({
    name   = string
    policy = optional(string, null)
  })
  description = "Configuration for the KMS key."
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
