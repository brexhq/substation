module "lambda_node" {
  source = "../../../../../../../build/terraform/aws/lambda"
  kms       = module.kms
  appconfig = module.appconfig

  config = {
    name        = "node"
    description = "Substation node that reads data from S3 via SNS."
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_HANDLER" : "AWS_S3_SNS"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
  ]
}

resource "aws_sns_topic_subscription" "node" {
  topic_arn = module.sns.arn
  protocol  = "lambda"
  endpoint  = module.lambda_node.arn
}

resource "aws_lambda_permission" "node" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_node.name
  principal     = "sns.amazonaws.com"
  source_arn    = module.sns.arn
}
