variable "name" {
  type = string
}

variable "kms_arn" {
  type = string
}

variable "tags" {
  type    = map(any)
  default = {}
}
