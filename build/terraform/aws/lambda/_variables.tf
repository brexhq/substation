variable "appconfig" {
  type = object({
    arn = string
    id = string
  })
}

variable "kms" {
  type = object({
    arn = string
    id = string
  })
}

variable "config" {
  type = object({
    name = string
    description = string
    image_uri = string
    architectures = optional(list(string), ["x86_64"])
    timeout = optional(number, 300)
    memory = optional(number, 1024)
    env = optional(map(any), null)
    secret = optional(bool, false)
    vpc_config = optional(object({
      subnet_ids = list(string)
      security_group_ids = list(string)
    }), null)
    iam_statements = optional(list(object({
      sid = string
      actions = list(string)
      resources = list(string)
    })), [])
  })
}

variable "tags" {
  type    = map(any)
  default = {}
}

variable "access" {
  type = list(string)
  default = []
  description = "List of IAM ARNs that are granted access to the resource."
}
