variable "config" {
  type = object({
    name        = string
    description = string
    function = object({
      arn  = string
      name = string
    })

    # Optional
    event_bus_arn = optional(string, null)
    event_pattern = optional(string, null)
    schedule      = optional(string, null)
  })
  description = <<EOH
    Configuration for the EventBridge Lambda rule:

    * name:         The name of the rule.
    * description:  The description of the rule.
    * function:     The Lambda function to invoke when the rule is triggered.
    * event_bus_arn: The ARN of the event bus to associate with the rule. If not provided, the default event bus is used.
    * event_pattern: The event pattern for the rule. If not provided, the rule is schedule-based.
    * schedule:     The schedule expression for the rule. If not provided, the rule is event-based.
EOH
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}

variable "access" {
  type        = list(string)
  default     = []
  description = "List of IAM ARNs that are granted access to the resource."
}
