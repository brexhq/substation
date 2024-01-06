variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  description = "KMS key used to encrypt the stream."
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
