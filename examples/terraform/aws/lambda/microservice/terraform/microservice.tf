module "microservice" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "microservice"
    description = "Substation node that acts as a synchronous microservice"
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true

    memory  = 128
    timeout = 10
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/microservice"
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
  function_name      = module.microservice.name
  authorization_type = "NONE"
}
