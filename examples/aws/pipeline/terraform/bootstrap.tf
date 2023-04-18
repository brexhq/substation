data "aws_caller_identity" "caller" {}

# KMS encryption key that is shared by all Substation infrastructure
module "kms_substation" {
  source = "../../../../build/terraform/aws/kms"
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

# use the dev environment for development resources
resource "aws_appconfig_environment" "dev" {
  name           = "dev"
  description    = "Stores development Substation configuration files"
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
  name    = "substation"
  kms_arn = module.kms_substation.arn
}

# repository for the autoscaling app
module "ecr_autoscaling" {
  source  = "../../../../build/terraform/aws/ecr"
  name    = "substation_autoscaling"
  kms_arn = module.kms_substation.arn
}

# repository for the validation app
module "ecr_validation" {
  source  = "../../../../build/terraform/aws/ecr"
  name    = "substation_validation"
  kms_arn = module.kms_substation.arn
}
