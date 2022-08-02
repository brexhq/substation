variable "name" {
  type = string
}

variable "stream" {
  type = string
}

# default timeout is 1 second / 1000 milliseconds
variable "timeout" {
  type    = number
  default = 1000
}

variable "tags" {
  type    = map(any)
  default = {}
}
