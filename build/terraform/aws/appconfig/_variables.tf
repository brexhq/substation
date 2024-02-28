variable "lambda" {
  type = object({
    name = string
    arn  = string
    role = object({
      name = string
      arn  = string
    })
  })
  default     = null
  description = "Lambda function used to validate configuration profiles."
}

variable "config" {
  type = object({
    name = string
    environments = list(object({
      name = string
    }))
  })

  description = <<EOH
    Configuration for the AppConfig application:

    * name: The name of the AppConfig application.
    * environments: A list of environments to create for the AppConfig application.
EOH
}

variable "tags" {
  type    = map(any)
  default = {}
}
