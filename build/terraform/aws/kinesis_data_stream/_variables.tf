variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "KMS key used to encrypt the stream. If not provided, then no server-side encryption is used. See https://docs.aws.amazon.com/streams/latest/dev/what-is-sse.html for more information."
}

variable "config" {
  type = object({
    name              = string
    autoscaling_topic = string
    shards            = optional(number, 2)
    retention         = optional(number, 24)
  })
  description = "Configuration for the Kinesis stream."
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
