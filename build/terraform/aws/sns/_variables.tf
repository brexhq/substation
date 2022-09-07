variable "name" {
  type = string
}

variable "kms_key_id" {
  type = string
}

variable "fifo" {
  type    = bool
  default = false
}

variable "tags" {
  type    = map(any)
  default = {}
}
