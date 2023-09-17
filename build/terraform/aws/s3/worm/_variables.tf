variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  description = "KMS key used to encrypt the bucket."
}

variable "config" {
  type = object({
    name          = string
    retention     = number
    force_destroy = optional(bool, false)
  })
  description = "Configuration for the S3 bucket."
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
