################################################
# appconfig permissions
# all Lambda must have this policy
################################################

module "iam_appconfig_read" {
  source    = "../../../../build/terraform/aws/iam"
  resources = ["${aws_appconfig_application.substation.arn}/*"]
}

module "iam_appconfig_read_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_appconfig_read"
  policy = module.iam_appconfig_read.appconfig_read_policy
  roles = [
    module.publisher.role,
    module.subscriber_x.role,
    module.subscriber_y.role,
    module.subscriber_z.role,
  ]
}
