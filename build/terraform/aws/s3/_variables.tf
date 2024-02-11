variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "Customer managed KMS key used to encrypt objects in the bucket. If not provided, then an S3 managed key is used. See https://docs.aws.amazon.com/AmazonS3/latest/userguide/serv-side-encryption.html for more information."
}

variable "config" {
  type = object({
    name          = string
    force_destroy = optional(bool, true)
    compliance = optional(object({
      retention = optional(number, 0)
    }))
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
