variable "config" {
  type = object({
    name = string
    autoscaling_topic = string
    shards = optional(number, 2)
    retention = optional(number, 24)
  })
}

variable kms {
  type = object({
    arn    = string
    id = string
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
