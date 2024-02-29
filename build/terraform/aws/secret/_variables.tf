variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "Customer managed KMS key used to encrypt the secret. If not provided, then an AWS managed key is used. See https://docs.aws.amazon.com/secretsmanager/latest/userguide/data-protection.html#encryption-at-rest for more information."
}

variable "config" {
  type = object({
    name = string
  })
  description = <<EOH
    Configuration for the secret:

    * name:    The name of the secret.
EOH
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
