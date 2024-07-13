variable "config" {
  type = object({
    name        = string
    description = string
    function = object({
      arn  = string
      name = string
    })
    # Optional
    schedule = optional(string, null)
    event = optional(object({
      bus_name = optional(string, null)
      pattern  = optional(string, null)
    }))
  })
  description = <<EOH
    Configuration for the EventBridge Lambda rule:

    * name:         The name of the rule.
    * description:  The description of the rule.
    * function:     The Lambda function to invoke when the rule is triggered.
    * schedule:     The schedule expression for the rule.
    * event:        The event route settings for the rule.
EOH
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
