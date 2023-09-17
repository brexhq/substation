variable "kms" {
  type = object({
    arn    = string
    id = string
  })
}

variable "config" {
  type = object({
    name = string
    hash_key = string
    attributes = list(object({
      name = string
      type = string
    }))

    range_key = optional(string, null)
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

    # change data capture via Streams is enabled by default for the table
    # https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html
    stream_view_type = optional(string, "NEW_AND_OLD_IMAGES")
  })
}

variable "tags" {
  type    = map(any)
  default = {}
}

variable "access" {
  type = list(string)
  default = []
  description = "List of IAM ARNs that are granted access to the resource."
}
