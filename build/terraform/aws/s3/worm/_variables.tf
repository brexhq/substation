variable "kms" {
  type = object({
    arn    = string
    id = string
  })
}

variable "config" {
  type = object({
    name = string
    retention = number
    force_destroy = optional(bool, false)
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
