variable "appconfig" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "AppConfig application used for configuring the function. If not provided, then no AppConfig configuration will be created for the function."
}

variable "kms" {
  type = object({
    arn = string
    id  = string
  })
  default     = null
  description = "Customer managed KMS key used to encrypt the function's environment variables. If not provided, then an AWS managed key is used. See https://docs.aws.amazon.com/lambda/latest/dg/security-dataprotection.html#security-privacy-atrest for more information."
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
  description = <<EOH
    Configuration for the Lambda function:

    * name: The name of the Lambda function.
    * description: The description of the Lambda function.
    * image_uri: The URI of the container image that contains the function code.
    * image_arm: Determines whether the image is an ARM64 image.
    * timeout: The amount of time that Lambda allows a function to run before stopping it. The default is 300 seconds.
    * memory: The amount of memory that your function has access to. The default is 1024 MB.
    * env: A map that defines environment variables for the function.
    * vpc_config: A map that defines the VPC configuration for the function.
    * iam_statements: A list of custom IAM policy statements to attach to the function's role.
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
