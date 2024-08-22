# Used for deploying and maintaining the Kinesis Data Streams autoscaling application; does not need to be used if deployments don't include Kinesis Data Streams.
module "lambda_autoscaling" {
  source = "../../../../../../build/terraform/aws/lambda"

  config = {
    name        = "autoscale"
    description = "Autoscaler for Kinesis Data Streams"
    image_uri   = "${module.ecr_autoscale.url}:v1.3.0"
    image_arm   = true
  }
}

# SNS topic for Kinesis Data Stream autoscaling alarms.
resource "aws_sns_topic" "autoscaling_topic" {
  name = "autoscale"
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
