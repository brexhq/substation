variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  description = "KMS key used to encrypt the table."
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

  description = "Configuration for the DynamoDB table."
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
