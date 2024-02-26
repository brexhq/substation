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

  description = <<EOH
    Configuration for the CloudWatch destination:

    * name: The name of the CloudWatch destination.
    * destination_arn: The ARN of the CloudWatch destination. This can be a Kinesis Data Firehose delivery stream or a Kinesis Data Streams stream.
    * account_ids: A list of AWS account IDs allowed to send events to the destination. If this is empty, then only the current account is allowed to send events to the destination.
EOH
}

variable "tags" {
  type    = map(any)
  default = {}
}
