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
    name    = "substation"
  }
}

# repository for the validation app
module "ecr_validation" {
  source  = "../../../../build/terraform/aws/ecr"
  kms = module.kms

  config = {
    name    = "substation_validation"
  }
}

module "sns" {
  source     = "../../../../build/terraform/aws/sns"
  kms   = module.kms

  config = {
    name = "substation"
  }

  access = [
    module.publisher.role,
    module.subscriber_x.role,
    module.subscriber_y.role,
    module.subscriber_z.role,
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
    module.publisher.role,
  ]
}
