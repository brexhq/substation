variable "bucket" {
  type = string
}

variable "kms_arn" {
  type = string
}

variable "retention_days" {
  type    = number
  default = 365
}

variable "tags" {
  type    = map(any)
  default = {}
}

variable "force_destroy" {
  type    = bool
  default = false
}
