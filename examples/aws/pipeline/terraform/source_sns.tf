module "lambda_source_sns" {
  source        = "../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "substation_source_sns"
    description = "Writes to Kinesis"
    image_uri = "${module.ecr_substation.url}:latest"
    architectures = ["arm64"]

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/prod/configurations/substation_source_sns"
      "SUBSTATION_HANDLER" : "AWS_SNS"
      "SUBSTATION_DEBUG" : true
    }
  }

  tags = {
    owner = "example"
  }
  
  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.url,
    module.vpc,
  ]
}

resource "aws_lambda_permission" "lambda_source_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_source_sns.name
  principal     = "sns.amazonaws.com"
  source_arn    = module.sns.arn

  depends_on = [
    module.lambda_source_sns.name
  ]
}

resource "aws_sns_topic_subscription" "lambda_source_sns" {
  topic_arn = module.sns.arn
  protocol  = "lambda"
  endpoint  = module.lambda_source_sns.arn

  depends_on = [
    module.lambda_source_sns.name
  ]
}
