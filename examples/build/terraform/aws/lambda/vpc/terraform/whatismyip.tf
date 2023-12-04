module "whatismyip" {
  source = "../../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "whatismyip"
    description = "Substation node that acts as a synchronous microservice"
    image_uri   = "${module.ecr_substation.url}:latest"
    image_arm   = true

    memory  = 128
    timeout = 10

    vpc_config = {
      subnet_ids         = module.vpc_substation.private_subnet_ids
      security_group_ids = [module.vpc_substation.default_security_group_id]
    }

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/whatismyip"
      "SUBSTATION_HANDLER" : "AWS_LAMBDA"
      "SUBSTATION_DEBUG" : true
    }
  }


  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.url,
  ]
}

resource "aws_lambda_function_url" "substation_microservice" {
  function_name      = module.whatismyip.name
  authorization_type = "NONE"
}
