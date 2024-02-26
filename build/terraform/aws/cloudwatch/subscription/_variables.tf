variable "config" {
  type = object({
    name            = string
    destination_arn = string
    log_groups      = list(string)
    filter_pattern  = optional(string, "")

  })

  description = <<EOH
    Configuration for the CloudWatch subscription filter:

    * name: The name of the CloudWatch subscription filter.
    * destination_arn: The ARN of the CloudWatch destination.
    * log_groups: The list of log groups to associate with the subscription filter.
    * filter_pattern: The filter pattern to use for the subscription filter. If not provided, all log events are sent to the destination.
EOH
}

variable "tags" {
  type    = map(any)
  default = {}
}
