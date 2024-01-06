data "aws_caller_identity" "caller" {}

# KMS encryption key that is shared by all Substation resources.
module "kms" {
  source = "../../../../../../../build/terraform/aws/kms"
  config = {
    name = "alias/substation"
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
  source = "../../../../../../../build/terraform/aws/ecr"
  kms    = module.kms

  config = {
    name         = "substation"
    force_delete = true
  }
}

# VPC shared by all Substation resources.
# 
# By default, this creates a /16 VPC with private subnets 
# in three availability zones in us-east-1.
module "vpc_substation" {
  source = "../../../../../../../build/terraform/aws/networking/vpc"
}
