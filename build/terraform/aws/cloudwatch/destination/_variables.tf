variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  description = "KMS key used to encrypt the resources."
}

variable "config" {
  type = object({
    name            = string
    destination_arn = string
    account_ids     = optional(list(string), [])
  })

  description = "Configuration for the CloudWatch destination."
}

variable "tags" {
  type    = map(any)
  default = {}
}
