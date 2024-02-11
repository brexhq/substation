variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "Customer managed KMS key used to encrypt messages in the queue. If not provided, then no server-side encryption is used. See https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html for more information."
}

variable "config" {
  type = object({
    name    = string
    delay   = optional(number, 0)
    timeout = optional(number, 30)
  })
  description = "Configuration for the SQS queue."

  validation {
    condition     = var.config.delay <= 900
    error_message = "Delay must be less than 15 minutes."
  }

  validation {
    condition     = var.config.timeout <= 43200
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
