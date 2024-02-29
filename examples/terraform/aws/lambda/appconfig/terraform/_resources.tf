module "appconfig" {
  source = "../../../../../../../build/terraform/aws/appconfig"
  lambda = module.validator

  config = {
    name = "substation"
    environments = [{
      name = "example"
    }]
  }
}

module "validator" {
  source = "../../../../../../../build/terraform/aws/lambda"
  config = {
    name        = "validator"
    description = "Substation validator that is executed by AppConfig."
    image_uri   = "${module.ecr_validator.url}:latest"
    image_arm   = true

    memory  = 128
    timeout = 1
  }

  depends_on = [
    module.appconfig.name,
    module.ecr_validator.url,
  ]
}

module "ecr" {
  source = "../../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}

module "ecr_validator" {
  source = "../../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation_validator"
    force_delete = true
  }
}
