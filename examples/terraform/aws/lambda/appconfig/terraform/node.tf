module "node" {
  source = "../../../../../../build/terraform/aws/lambda"

  # AppConfig is configured to validate configurations before deployment.
  appconfig = module.appconfig

  config = {
    name        = "node"
    description = "Substation node that never receives a configuration."
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true

    memory  = 128
    timeout = 10
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_LAMBDA"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
  ]
}

resource "aws_lambda_function_url" "url" {
  function_name      = module.node.name
  authorization_type = "NONE"
}
