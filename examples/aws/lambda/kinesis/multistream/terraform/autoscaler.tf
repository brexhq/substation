# Used for deploying and maintaining the Kinesis Data Streams autoscaling application; does not need to be used if deployments don't include Kinesis Data Streams.

module "lambda_autoscaling" {
  source = "../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "autoscaler"
    description = "Autoscaler for Kinesis Data Streams"
    image_uri   = "${module.ecr_autoscaling.url}:latest"
    image_arm   = true
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.url,
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
