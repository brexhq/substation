variable "config" {
  type = object({
    name = string
    function = object({
      arn = string
    })
  })
}

variable "tags" {
  type    = map(any)
  default = {}
}
