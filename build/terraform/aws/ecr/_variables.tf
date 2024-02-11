variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "Customer managed KMS key used to encrypt the image. If not provided, then an S3 managed key is used. See https://docs.aws.amazon.com/AmazonECR/latest/userguide/encryption-at-rest.html for more information."
}

variable "config" {
  type = object({
    name         = string
    force_delete = optional(bool, false)
  })
  description = "Configuration for the ECR image."
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
