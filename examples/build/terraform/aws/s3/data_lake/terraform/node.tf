module "lambda_gateway" {
  source = "../../../../../../../build/terraform/aws/api_gateway/lambda"
  lambda = module.lambda_node

  config = {
    name = "node_gateway"
  }

  depends_on = [
    module.lambda_node
  ]
}

module "lambda_node" {
  source = "../../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "node"
    description = "Substation node that writes data to S3"
    image_uri   = "${module.ecr_substation.url}:latest"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_HANDLER" : "AWS_API_GATEWAY"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.url,
  ]
}
