data "aws_caller_identity" "caller" {}

# KMS encryption key that is shared by all Substation resources.
module "kms" {
  source = "../../../../../build/terraform/aws/kms"
  config = {
    name   = "alias/substation"
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

# AppConfig application that is shared by all Substation applications.
resource "aws_appconfig_application" "substation" {
  name        = "substation"
  description = "Stores compiled configuration files for Substation"
}

resource "aws_appconfig_environment" "example" {
  name           = "example"
  description    = "Stores example Substation configuration files"
  application_id = aws_appconfig_application.substation.id
}

# AWS Lambda requires an instant deployment strategy.
resource "aws_appconfig_deployment_strategy" "instant" {
  name                           = "Instant"
  description                    = "This strategy deploys the configuration to all targets immediately with zero bake time."
  deployment_duration_in_minutes = 0
  final_bake_time_in_minutes     = 0
  growth_factor                  = 100
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"
}

# Repository for the core Substation application.
module "ecr_substation" {
  source = "../../../../../build/terraform/aws/ecr"
  kms    = module.kms

  config = {
    name = "substation"
  }
}

module "sqs" {
  source = "../../../../../build/terraform/aws/sqs"
  kms    = module.kms

  config = {
    name = "substation"
  }

  access = [
    # Reads from SQS.
    module.microservice.role.name,
    # Writes to SQS.
    module.frontend.role.name,
  ]
}

module "dynamodb" {
  source = "../../../../../build/terraform/aws/dynamodb"
  kms    = module.kms

  config = {
    name     = "substation"
    hash_key = "PK"
    attributes = [
      {
        name = "PK"
        type = "S"
      }
    ]
  }

  access = [
    # Writes to DynamoDB.
    module.microservice.role.name,
  ]
}
