module "node" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "node"
    description = "Substation node that transforms data exactly-once using a distributed lock"
    image_uri   = "${module.ecr.url}:v1.3.0"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_API_GATEWAY"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
  ]
}

resource "aws_lambda_function_url" "node" {
  function_name      = module.node.name
  authorization_type = "NONE"
}
