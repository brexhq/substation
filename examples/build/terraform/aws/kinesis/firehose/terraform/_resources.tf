locals {
  name = "firehose"
}

data "aws_caller_identity" "caller" {}

module "kms" {
  source = "../../../../../../../build/terraform/aws/kms"

  config = {
    name = "alias/substation"
  }
}

module "appconfig" {
  source = "../../../../../../../build/terraform/aws/appconfig"

  config = {
    name        = "substation"
    environments = [{
      name = "example"
    }]
  }
}

module "ecr" {
  source = "../../../../../../../build/terraform/aws/ecr"
  kms    = module.kms

  config = {
    name         = "substation"
    force_delete = true
  }
}

##################################
# Kinesis Data Firehose resources
##################################

# IAM
data "aws_iam_policy_document" "firehose" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["firehose.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "firehose" {
  name               = "sub-${local.name}"
  assume_role_policy = data.aws_iam_policy_document.firehose.json
}


data "aws_iam_policy_document" "firehose_s3" {
  statement {
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]

    resources = [
      module.kms.arn,
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "s3:AbortMultipartUpload",
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket",
      "s3:ListBucketMultipartUploads",
      "s3:PutObject"
    ]

    resources = [
      aws_s3_bucket.firehose_s3.arn,
      "${aws_s3_bucket.firehose_s3.arn}/*",
    ]
  }
}

resource "aws_iam_policy" "firehose_s3" {
  name        = "sub-${local.name}"
  description = "Policy for the ${local.name} Kinesis Data Firehose."
  policy      = data.aws_iam_policy_document.firehose_s3.json
}


resource "aws_iam_role_policy_attachment" "firehose_s3" {
  role       = aws_iam_role.firehose.name
  policy_arn = aws_iam_policy.firehose_s3.arn
}

# S3
resource "random_uuid" "firehose_s3" {}

resource "aws_s3_bucket" "firehose_s3" {
  bucket = "${random_uuid.firehose_s3.result}-substation"

}

resource "aws_s3_bucket_ownership_controls" "firehose_s3" {
  bucket = aws_s3_bucket.firehose_s3.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "firehose_s3" {
  bucket = aws_s3_bucket.firehose_s3.id
  acl    = "private"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "firehose_s3" {
  bucket = aws_s3_bucket.firehose_s3.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = module.kms.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

# Kinesis Data Firehose
resource "aws_kinesis_firehose_delivery_stream" "firehose" {
  name        = "substation"
  destination = "extended_s3"

  server_side_encryption {
    enabled  = true
    key_type = "CUSTOMER_MANAGED_CMK"
    key_arn  = module.kms.arn
  }

  extended_s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    bucket_arn         = aws_s3_bucket.firehose_s3.arn
    kms_key_arn        = module.kms.arn
    buffering_interval = 60

    processing_configuration {
      enabled = "true"

      processors {
        type = "Lambda"

        parameters {
          parameter_name = "LambdaArn"
          # LATEST is always used for container images.
          parameter_value = "${module.processor.arn}:$LATEST"
        }
      }
    }
  }
}

module "processor" {
  source = "../../../../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = module.appconfig

  config = {
    name        = "processor"
    description = "Processes Kinesis Data Firehose records."
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true

    memory  = 128
    timeout = 60
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/processor"
      "SUBSTATION_HANDLER" : "AWS_KINESIS_DATA_FIREHOSE"
      "SUBSTATION_DEBUG" : true
    }
  }

  access = [
    aws_iam_role.firehose.name,
  ]

  depends_on = [
    module.appconfig.name,
    module.ecr_substation.url,
  ]
}
