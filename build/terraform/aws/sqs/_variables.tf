variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  description = "KMS key used to encrypt the queue."
}

variable "config" {
  type = object({
    name    = string
    delay   = optional(number, 0)
    timeout = optional(number, 30)
  })
  description = "Configuration for the SQS queue."

  validation {
    condition     = var.config.delay > 900
    error_message = "Delay must be less than 15 minutes."
  }

  validation {
    condition     = var.config.timeout > 43200
    error_message = "Timeout must be less than 12 hours."
  }
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}

variable "access" {
  type        = list(string)
  default     = []
  description = "List of IAM ARNs that are granted access to the resource."
}
