# S3 bucket shared by all resources
resource "random_uuid" "bucket" {}

module "s3_substation" {
  source  = "../../../../build/terraform/aws/s3"
  kms_arn = module.kms_substation.arn
  bucket  = "${random_uuid.bucket.result}-substation"
}

################################################
## bucket read permissions
################################################

module "iam_s3_read" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    "${module.s3_substation.arn}/*",
  ]
}

module "iam_s3_read_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "iam_s3_read"
  policy = module.iam_s3_read.s3_read_policy
  roles = [
    module.lambda_s3_source.role,
  ]
}

################################################
## bucket write permissions
################################################


module "iam_s3_write" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    "${module.s3_substation.arn}/*",
  ]
}

module "iam_s3_write_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "iam_s3_write"
  policy = module.iam_s3_write.s3_write_policy
  roles = [
    module.lambda_raw_s3_sink.role,
    module.lambda_processed_s3_sink.role,
  ]
}
