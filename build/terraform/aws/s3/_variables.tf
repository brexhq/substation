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
  description = <<EOH
    Configuration for the S3 bucket:

    * name:    The name of the bucket.
    * force_destroy: A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are not recoverable.
    * compliance.retention: The default retention period for objects in the bucket. The value is in days. **Note: this enables Compliance mode for objects in the bucket.** See https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html#object-lock-retention-modes for more information.
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
