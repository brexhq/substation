variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  description = "KMS key used to encrypt the repository."
}

variable "config" {
  type = object({
    name         = string
    force_delete = optional(bool, false)
  })
  description = "Configuration for the ECR repository."
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
