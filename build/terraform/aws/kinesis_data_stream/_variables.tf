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
  description = <<EOH
    Configuration for the Kinesis Data Stream:

    * name: The name of the Kinesis Data Stream.
    * autoscaling_topic: The ARN of the SNS topic that will be used for autoscaling.
    * shards: The number of shards to create for the stream. Defaults to 2.
    * retention: The number of hours to retain data records in the stream. Defaults to 24.
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
