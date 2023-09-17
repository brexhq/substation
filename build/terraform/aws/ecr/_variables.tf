variable "config" {
  type = object({
    name = string
  })  
}

variable "kms" {
  type = object({
    arn    = string
    id = string
  })
}

variable "tags" {
  type    = map(any)
  default = {}
}
