variable "function_name" {
  type = string
}

variable "description" {
  type = string
}

variable "image_uri" {
  type = string
}

variable "appconfig_id" {
  type = string
}

variable "architectures" {
  type    = list(string)
  default = ["x86_64"]
}

variable "timeout" {
  type    = number
  default = 300
}

variable "memory_size" {
  type    = number
  default = 1024
}

variable "env" {
  type    = map(any)
  default = null
}

variable "kms_arn" {
  type = string
}

variable "tags" {
  type    = map(any)
  default = {}
}

variable "secret" {
  type    = bool
  default = false
}

variable "vpc_config" {
  type = object({
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
  default = {
    subnet_ids         = []
    security_group_ids = []
  }
}
