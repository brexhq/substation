# Used for deploying and maintaining the Kinesis Data Streams autoscaling application; does not need to be used if deployments don't include Kinesis Data Streams.

resource "aws_sns_topic" "autoscaling_topic" {
  name              = "substation_autoscaling"
  kms_master_key_id = module.kms.id

  tags = {
    owner = "example"
  }
}

module "lambda_autoscaling" {
  source        = "../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "substation_autoscaling"
    description = "Autoscales Kinesis streams based on data volume and size"
    image_uri = "${module.ecr_autoscaling.url}:latest"
    architectures = ["arm64"]
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

resource "aws_sns_topic_subscription" "autoscaling_subscription" {
  topic_arn = aws_sns_topic.autoscaling_topic.arn
  protocol  = "lambda"
  endpoint  = module.lambda_autoscaling.arn

  depends_on = [
    module.lambda_autoscaling.name
  ]
}

resource "aws_lambda_permission" "autoscaling_invoke" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_autoscaling.name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.autoscaling_topic.arn

  depends_on = [
    module.lambda_autoscaling.name
  ]
}
