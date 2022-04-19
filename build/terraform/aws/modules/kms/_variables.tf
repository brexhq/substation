variable "name" {
  type = string
}

variable "tags" {
  type    = map(any)
  default = {}
}

variable "policy" {
  type    = string
  default = ""
}
