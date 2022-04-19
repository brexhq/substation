variable "name" {
  type = string
}

variable "function_arn" {
  type = string
}

variable "tags" {
  type    = map(any)
  default = {}
}
