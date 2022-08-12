variable "name" {
  type = string
}

variable "kms_key_id" {
  type = string
}

variable "delay_seconds" {
  type    = number
  default = 0
}

variable "visibility_timeout_seconds" {
  type    = number
  default = 300
}

variable "fifo" {
  type    = bool
  default = false
}

variable "tags" {
  type    = map(any)
  default = {}
}
