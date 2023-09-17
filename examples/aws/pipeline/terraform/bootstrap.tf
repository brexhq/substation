data "aws_caller_identity" "caller" {}

# KMS encryption key that is shared by all Substation infrastructure
module "kms" {
  source = "../../../../build/terraform/aws/kms"

  config = {
    name = "alias/substation"
    policy = <<POLICY
  {
    "Version": "2012-10-17",
    "Statement": [
      {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Principal": {
        "Service": "cloudwatch.amazonaws.com"
      },
      "Resource": "*"
      },
      {
      "Effect": "Allow",
      "Action": "kms:*",
      "Principal": {
        "AWS": "arn:aws:iam::${data.aws_caller_identity.caller.account_id}:root"
      },
      "Resource": "*"
      }
    ]
  }
  POLICY
  }
}

# AppConfig application that is shared by all Substation apps
resource "aws_appconfig_application" "substation" {
  name        = "substation"
  description = "Stores compiled configuration files for Substation"
}

# use the prod environment for production resources
resource "aws_appconfig_environment" "prod" {
  name           = "prod"
  description    = "Stores production Substation configuration files"
  application_id = aws_appconfig_application.substation.id
}

# AppConfig doesn't have useful support for non-linear, non-instant deployments on AWS Lambda, so this deployment strategy is used to deploy configurations as quickly as possible
# todo: add configuration rollback via CloudWatch Lambda monitoring
resource "aws_appconfig_deployment_strategy" "instant" {
  name                           = "Instant"
  description                    = "This strategy deploys the configuration to all targets immediately with zero bake time."
  deployment_duration_in_minutes = 0
  final_bake_time_in_minutes     = 0
  growth_factor                  = 100
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"
}

# repository for the core Substation app
module "ecr_substation" {
  source  = "../../../../build/terraform/aws/ecr"
  kms = module.kms

  config = {
    name = "substation"
  }
}

# repository for the autoscaling app
module "ecr_autoscaling" {
  source  = "../../../../build/terraform/aws/ecr"
  kms = module.kms

  config = {
    name = "substation_autoscaling"
  }
}

# repository for the validation app
module "ecr_validation" {
  source  = "../../../../build/terraform/aws/ecr"
  kms = module.kms

  config = {
    name = "substation_validation"
  }
}

# By default this creates these resources:
# - Address space of 64k hosts (split between all subnets)
# - 1 public subnet with an internet gateway
# - 3 private subnets each with a NAT gateway
module "vpc" {
  source = "../../../../build/terraform/aws/networking/vpc"

  config = {}
}

module "gateway_to_kinesis" {
  source = "../../../../build/terraform/aws/api_gateway/kinesis_data_stream"

  config = {
    name   = "substation_kinesis_gateway"
    stream = "substation_raw"
  }
}

module "kinesis_raw" {
  source            = "../../../../build/terraform/aws/kinesis_data_stream"
  kms = module.kms

  config = {
    name = "substation_raw"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
  }

  tags = {
    owner = "example"
  }

  access = [
    module.lambda_autoscaling.role,
    module.lambda_processor.role,
    module.lambda_sink_s3.role,
  ]
}


module "kinesis_processed" {
  source            = "../../../../build/terraform/aws/kinesis_data_stream"
  kms        = module.kms

  config = {
    name = "substation_processed"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
  }

  tags = {
    owner = "example"
  }

  access = [
    module.lambda_autoscaling.role,
    module.lambda_processor.role,
    module.lambda_sink_s3.role,
    module.lambda_source_sns.role,
  ]
}

resource "random_uuid" "s3" {}

module "s3" {
  source  = "../../../../build/terraform/aws/s3"
  kms = module.kms

  config = {
    name    = "${random_uuid.s3.result}-substation"
  }

  access = [
    module.lambda_source_s3.role,
    module.lambda_sink_s3.role,
  ]
}

module "dynamodb" {
  source     = "../../../../build/terraform/aws/dynamodb"
  kms    = module.kms

  config = {
    name = "substation"
    hash_key = "PK"
    attributes = [
      {
        name = "PK"
        type = "S"
      }
    ]
  }

  access = [
    module.lambda_sink_dynamodb.role,
  ]
}

module "sns" {
  source     = "../../../../build/terraform/aws/sns"
  kms = module.kms

  config = {
    name = "substation"
  }

  access = [
    module.lambda_source_sns.role,
  ]
}

module "sqs" {
  source     = "../../../../build/terraform/aws/sqs"
  kms = module.kms
  
  config = {
    name = "substation"
    # Timeout must match timeout on Lambda.
    timeout = 300
  }

  access = [
    module.lambda_source_sqs.role,
  ]
}
