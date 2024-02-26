variable "lambda" {
  type = object({
    arn = string
  })
  description = "Lambda invoked by the API Gateway."
}

variable "config" {
  type = object({
    name = string
  })

  description = <<EOH
    Configuration for the API Gateway Lambda integration:

    * name: The name of the API Gateway.
EOH
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
