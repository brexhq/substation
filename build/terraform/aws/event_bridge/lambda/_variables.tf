variable "name" {
  type = string
}

variable "description" {
  type = string
}

variable "schedule_expression" {
  type = string
}

variable "function_arn" {
  type = string
}

variable "function_name" {
  type = string
}

variable "tags" {
  type    = map(any)
  default = {}
}
