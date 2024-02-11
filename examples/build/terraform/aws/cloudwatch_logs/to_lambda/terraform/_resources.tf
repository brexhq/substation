data "aws_caller_identity" "caller" {}

module "appconfig" {
  source = "../../../../../../../build/terraform/aws/appconfig"

  config = {
    name = "substation"
    environments = [{
      name = "example"
    }]
  }
}

module "ecr" {
  source = "../../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}
