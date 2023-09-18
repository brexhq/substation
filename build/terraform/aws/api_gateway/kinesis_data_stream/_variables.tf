variable "kinesis_data_stream" {
  type = object({
    name = string
  })
  description = "Kinesis Data Stream requests are sent to."
}

variable "config" {
  type = object({
    name    = string
    timeout = optional(number, 1000)
  })
  description = "Configuration for the API Gateway Kinesis Data Stream integration."
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
