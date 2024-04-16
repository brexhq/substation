module "lambda_autoscale" {
  source    = "../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "autoscale"
    description = "Autoscaler for Kinesis Data Streams"
    image_uri   = "${module.ecr_autoscale.url}:v1.2.0"
    image_arm   = true
  }
}

resource "aws_sns_topic_subscription" "autoscaling_subscription" {
  topic_arn = aws_sns_topic.autoscaling_topic.arn
  protocol  = "lambda"
  endpoint  = module.lambda_autoscale.arn

  depends_on = [
    module.lambda_autoscale.name
  ]
}

resource "aws_lambda_permission" "autoscaling_invoke" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_autoscale.name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.autoscaling_topic.arn

  depends_on = [
    module.lambda_autoscale.name
  ]
}
