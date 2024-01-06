variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  description = "KMS key used to encrypt the secret."
}

variable "config" {
  type = object({
    name = string
  })
  description = "Configuration for the secret."
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
