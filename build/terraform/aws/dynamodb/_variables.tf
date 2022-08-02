variable "kms_arn" {
  type = string
}

variable "table_name" {
  type = string
}

variable "read_capacity_min" {
  type    = number
  default = 5
}

variable "read_capacity_max" {
  type    = number
  default = 1000
}

variable "read_capacity_target" {
  type    = number
  default = 70
}

variable "write_capacity_min" {
  type    = number
  default = 5
}

variable "write_capacity_max" {
  type    = number
  default = 1000
}

variable "write_capacity_target" {
  type    = number
  default = 70
}

variable "hash_key" {
  type = string
}

variable "range_key" {
  type    = string
  default = null
}

variable "attributes" {
  type = list(object({
    name = string
    type = string
  }))
}

variable "tags" {
  type    = map(any)
  default = {}
}
