data "aws_caller_identity" "caller" {}

module "appconfig" {
  source = "../../../../../../build/terraform/aws/appconfig"

  config = {
    name = "substation"
    environments = [{
      name = "example"
    }]
  }
}

# Repository for the core Substation application.
module "ecr" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}

resource "random_uuid" "s3" {}

# Monolithic S3 bucket used to store all data.
module "s3" {
  source = "../../../../../../build/terraform/aws/s3"

  config = {
    # Bucket name is randomized to avoid collisions.
    name = "substation-${random_uuid.s3.result}"
  }

  access = [
    module.lambda_node.role.name,
  ]
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = module.s3.id

  lambda_function {
    lambda_function_arn = module.lambda_node.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "data/"
  }

  depends_on = [aws_lambda_permission.allow_bucket]
}
