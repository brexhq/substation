variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "Customer managed KMS key used to encrypt messages in the topic. If not provided, then no server-side encryption is used. See https://docs.aws.amazon.com/sns/latest/dg/sns-server-side-encryption.html for more information."
}

variable "config" {
  type = object({
    name = string
  })
  description = "Configuration for the SNS topic."
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
