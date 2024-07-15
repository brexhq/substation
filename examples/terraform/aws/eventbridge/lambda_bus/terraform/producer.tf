module "eventbridge_producer" {
  source = "../../../../../../build/terraform/aws/eventbridge/lambda"

  config = {
    name        = "substation_producer"
    description = "Sends messages to the default EventBridge bus on a schedule."
    function    = module.lambda_producer # This is the Lambda function that will be invoked.
    schedule    = "rate(1 minute)"
  }
}

module "lambda_producer" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "producer"
    description = "Substation node that is invoked by the EventBridge schedule."
    image_uri   = "${module.ecr.url}:v1.5.0"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/producer"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_LAMBDA"
      "SUBSTATION_DEBUG" : true
    }
  }
}
