variable "config" {
  type = object({
    name        = string
    description = string
    schedule    = string
    function = object({
      arn  = string
      name = string
    })
  })
  description = "Configuration for the EventBridge Lambda rule."
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
