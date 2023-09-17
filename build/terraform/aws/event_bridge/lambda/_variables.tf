variable "config" {
  type = object({
    name = string
    description = string
    schedule = string
    function = object({
      arn = string
      name = string
    })
  })  
}

variable "tags" {
  type    = map(any)
  default = {}
}
