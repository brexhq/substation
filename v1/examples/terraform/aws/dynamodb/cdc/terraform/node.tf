module "node" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "node"
    description = "Substation node that receives CDC events"
    image_uri   = "${module.ecr.url}:v1.3.0"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_DYNAMODB_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
  ]
}

resource "aws_lambda_event_source_mapping" "node" {
  event_source_arn  = module.dynamodb.stream_arn
  function_name     = module.node.arn
  starting_position = "LATEST"
}
