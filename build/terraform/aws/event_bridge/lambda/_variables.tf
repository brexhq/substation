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
  description = <<EOH
    Configuration for the EventBridge Lambda rule:

    * name:         The name of the rule.
    * description:  The description of the rule.
    * schedule:     The schedule expression for the rule.
    * function:     The Lambda function to invoke when the rule is triggered.
EOH
}

variable "tags" {
  type        = map(any)
  default     = {}
  description = "Tags to apply to all resources."
}
