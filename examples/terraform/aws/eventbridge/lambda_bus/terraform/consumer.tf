module "eventbridge_consumer" {
  source = "../../../../../../build/terraform/aws/eventbridge/lambda"

  config = {
    name        = "substation_consumer"
    description = "Routes messages from any Substation producer to a Substation Lambda consumer."
    function    = module.lambda_consumer # This is the Lambda function that will be invoked.
    event_pattern = jsonencode({
      # This matches every event sent by any Substation app.
      source = [{ "wildcard" : "substation.*" }]
    })
  }

  access = [
    module.lambda_producer.role.name,
  ]
}

module "lambda_consumer" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "consumer"
    description = "Substation node that is invoked by the EventBridge bus."
    image_uri   = "${module.ecr.url}:v1.5.0"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/consumer"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_LAMBDA"
      "SUBSTATION_DEBUG" : true
    }
  }
}
