################################################
# KMS read permissions
# all Lambda must have this policy
################################################

module "iam_kms_read" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.kms.arn,
  ]
}

module "iam_kms_read_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_kms_read"
  policy = module.iam_kms_read.kms_read_policy
  roles = [
    module.publisher.role,
    module.subscriber_x.role,
    module.subscriber_y.role,
    module.subscriber_z.role,
  ]
}

################################################
# KMS write permissions
# all Lambda must have this policy
################################################

module "iam_kms_write" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.kms.arn,
  ]
}

module "iam_kms_write_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_kms_write"
  policy = module.iam_kms_write.kms_write_policy
  roles = [
    module.publisher.role,
    module.subscriber_x.role,
    module.subscriber_y.role,
    module.subscriber_z.role,
  ]
}
