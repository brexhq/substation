################################################
# Substation node that publishes CDC events to SNS
################################################

module "publisher" {
  source        = "../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms   = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "publisher"
    description = "Publishes CDC events to SNS"
    image_uri = "${module.ecr_substation.url}:latest"
    architectures = ["arm64"]
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/prod/configurations/publisher"
      "SUBSTATION_HANDLER" : "AWS_DYNAMODB_STREAM"
      "SUBSTATION_DEBUG" : true
    }
  }

  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.url,
  ]
}

resource "aws_lambda_event_source_mapping" "publisher" {
  event_source_arn  = module.dynamodb.stream_arn
  function_name     = module.publisher.arn
  starting_position = "LATEST"
}
