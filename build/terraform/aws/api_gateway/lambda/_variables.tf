variable "config" {
  type = object({
    name = string
    function = object({
      arn = string
    })
  })
  description = "Configuration for the API Gateway Lambda integration."
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
