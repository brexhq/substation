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

module "ecr" {
  source = "../../../../../../build/terraform/aws/ecr"

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
  source = "../../../../../../build/terraform/aws/networking/vpc"
}
