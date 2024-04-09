module "frontend" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "frontend"
    description = "Substation node that acts as a frontend to an asynchronous microservice"
    image_uri   = "${module.ecr.url}:v1.2.0"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/frontend"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_LAMBDA"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
  ]
}

resource "aws_lambda_function_url" "frontend" {
  function_name      = module.frontend.name
  authorization_type = "NONE"
}
