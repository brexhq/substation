variable "bucket" {
  type = string
}

variable "kms_arn" {
  type = string
}

variable "tags" {
  type    = map(any)
  default = {}
}

variable "force_destroy" {
  type    = bool
  default = true
}
