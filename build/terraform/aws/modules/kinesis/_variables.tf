variable "stream_name" {
  type = string
}

variable "kms_key_id" {
  type = string
}

variable "autoscaling_topic" {
  type = string
}

variable "shard_count" {
  type    = number
  default = 1
}

variable "retention_period" {
  type    = number
  default = 24
}

variable "tags" {
  type    = map(any)
  default = {}
}
