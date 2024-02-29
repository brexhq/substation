variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "Customer managed KMS key used to encrypt the table. If not provided, then an AWS owned key is used. See https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/EncryptionAtRest.html for more information."
}

variable "config" {
  type = object({
    name     = string
    hash_key = string
    attributes = list(object({
      name = string
      type = string
    }))

    range_key = optional(string, null)
    ttl       = optional(string, null)
    read_capacity = optional(object({
      min    = optional(number, 5)
      max    = optional(number, 1000)
      target = optional(number, 70)
    }))
    write_capacity = optional(object({
      min    = optional(number, 5)
      max    = optional(number, 1000)
      target = optional(number, 70)
    }))

    # Change Data Capture via Streams is enabled by default for the table.
    # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html
    stream_view_type = optional(string, "NEW_AND_OLD_IMAGES")
  })

  description = <<EOH
    Configuration for the DynamoDB table:

    * name:         The name of the table.
    * hash_key:     The name of the hash key (aka Partition Key).
    * range_key:    The name of the range key (aka Sort Key).
    * ttl:          The name of the attribute to use for TTL.
    * attributes:   A list of attributes for the table. The first attribute is the hash key, and the second is the range key.
    * read_capacity:  The read capacity settings for the table.
    * write_capacity: The write capacity settings for the table.
    * stream_view_type: The type of data from the table to be written to the stream. Valid values are NEW_IMAGE, OLD_IMAGE, NEW_AND_OLD_IMAGES, and KEYS_ONLY. The default value is NEW_AND_OLD_IMAGES. See https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_StreamSpecification.html for more information.
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
