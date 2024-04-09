module "whatismyip" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "whatismyip"
    description = "Substation node that acts as a synchronous microservice"
    image_uri   = "${module.ecr.url}:v1.2.0"
    image_arm   = true

    memory  = 128
    timeout = 10

    vpc_config = {
      subnet_ids         = module.vpc_substation.private_subnet_ids
      security_group_ids = [module.vpc_substation.default_security_group_id]
    }

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/whatismyip"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_LAMBDA"
      "SUBSTATION_DEBUG" : true
    }
  }


  depends_on = [
    module.appconfig.name,
    module.ecr.url,
  ]
}

resource "aws_lambda_function_url" "substation_microservice" {
  function_name      = module.whatismyip.name
  authorization_type = "NONE"
}
