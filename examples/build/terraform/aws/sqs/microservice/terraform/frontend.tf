module "frontend" {
  source = "../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "frontend"
    description = "Substation node that acts as a frontend to an asynchronous microservice"
    image_uri   = "${module.ecr_substation.url}:latest"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/frontend"
      "SUBSTATION_HANDLER" : "AWS_LAMBDA"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.url,
  ]
}

resource "aws_lambda_function_url" "frontend" {
  function_name      = module.frontend.name
  authorization_type = "NONE"
}
