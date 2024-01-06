module "node" {
  source = "../../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "node"
    description = "Substation node that receives CDC events"
    image_uri   = "${module.ecr_substation.url}:latest"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_HANDLER" : "AWS_DYNAMODB_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.url,
  ]
}

resource "aws_lambda_event_source_mapping" "node" {
  event_source_arn  = module.dynamodb.stream_arn
  function_name     = module.node.arn
  starting_position = "LATEST"
}
