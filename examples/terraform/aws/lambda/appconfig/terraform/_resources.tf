module "appconfig" {
  source = "../../../../../../build/terraform/aws/appconfig"
  lambda = module.validate

  config = {
    name = "substation"
    environments = [{
      name = "example"
    }]
  }
}

module "validate" {
  source = "../../../../../../build/terraform/aws/lambda"
  config = {
    name        = "validate"
    description = "Substation configuration validator that is executed by AppConfig."
    image_uri   = "${module.ecr_validate.url}:v1.3.0"
    image_arm   = true

    memory  = 128
    timeout = 1
  }

  depends_on = [
    module.appconfig.name,
    module.ecr_validate.url,
  ]
}

module "ecr" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}

module "ecr_validate" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "validate"
    force_delete = true
  }
}
