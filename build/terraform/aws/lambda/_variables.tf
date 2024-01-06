variable "appconfig" {
  type = object({
    arn = string
    id  = string
  })
  description = "AppConfig application used for the Lambda configuration."
}

variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  description = "KMS key used to encrypt the Lambda."
}

variable "config" {
  type = object({
    name        = string
    description = string
    image_uri   = string
    image_arm   = bool
    timeout     = optional(number, 300)
    memory      = optional(number, 1024)
    env         = optional(map(any), null)
    vpc_config = optional(object({
      subnet_ids         = list(string)
      security_group_ids = list(string)
      }), {
      subnet_ids         = []
      security_group_ids = []
    })
    iam_statements = optional(list(object({
      sid       = string
      actions   = list(string)
      resources = list(string)
    })), [])
  })
  description = "Configuration for the Lambda function."
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
