variable "config" {
  type = object({
    name        = string
    description = string
    environments = list(object({
      name = string
    }))
  })

  description = "Configuration for the AppConfig application."
}

variable "tags" {
  type    = map(any)
  default = {}
}
