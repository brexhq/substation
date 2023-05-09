variable "name" {
  type = string
}

variable "kms_key_id" {
  type = string
}

variable "tags" {
  type    = map(any)
  default = {}
}
