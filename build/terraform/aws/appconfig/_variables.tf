variable "config" {
  type = object({
    name = string
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
