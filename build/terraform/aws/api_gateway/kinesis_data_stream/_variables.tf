variable "config" {
  type = object({
    name    = string
    stream  = string
    timeout = optional(number, 1000)
  })
  description = "Configuration for the API Gateway Kinesis Data Stream integration."
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
