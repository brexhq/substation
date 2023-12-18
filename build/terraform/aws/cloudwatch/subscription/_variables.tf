variable "config" {
  type = object({
    name            = string
    destination_arn = string
    log_groups      = list(string)
    filter_pattern  = optional(string, "")

  })

  description = "Configuration for the CloudWatch subscription filter."
}

variable "tags" {
  type    = map(any)
  default = {}
}
