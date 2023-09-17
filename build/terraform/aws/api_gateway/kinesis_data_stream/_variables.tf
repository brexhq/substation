variable "config" {
  type = object({
    name = string
    stream = string
    timeout = optional(number, 1000)
  })  
}

variable "tags" {
  type    = map(any)
  default = {}
}
