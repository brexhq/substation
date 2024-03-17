# Repository for the Autoscale app.
module "ecr" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "autoscale"
    force_delete = true
  }
}

# SNS topic for Kinesis Data Stream autoscale alarms.
resource "aws_sns_topic" "autoscale" {
  name = "autoscale"
}

# Kinesis Data Stream that is managed by the Autoscale app.
module "kds" {
  source = "../../../../../../build/terraform/aws/kinesis_data_stream"

  config = {
    name      = "substation"
    autoscale = aws_sns_topic.autoscale.arn
  }

  # Add additional consumer and producer roles as needed.
  access = [
    # Autoscales the stream.
    module.lambda_autoscale.role.name,
  ]
}

# Lambda Autoscale application that manages Kinesis Data Streams.
module "lambda_autoscale" {
  source = "../../../../../../build/terraform/aws/lambda"

  config = {
    name        = "autoscale"
    description = "Autoscaler for Kinesis Data Streams."
    image_uri   = "${module.ecr.url}:latest" # This should use the project's release tags.
    image_arm   = true

    # Override the default Autoscale configuration using environment variables.
    # These are the default settings, included for demonstration purposes.
    env = {
      "AUTOSCALE_KINESIS_THRESHOLD" : 0.7,
      "AUTOSCALE_KINESIS_UPSCALE_DATAPOINTS" : 5,
      "AUTOSCALE_KINESIS_DOWNSCALE_DATAPOINTS" : 60,
    }
  }

  depends_on = [
    module.ecr.url,
  ]
}

resource "aws_sns_topic_subscription" "autoscale_subscription" {
  topic_arn = aws_sns_topic.autoscale.arn
  protocol  = "lambda"
  endpoint  = module.lambda_autoscale.arn

  depends_on = [
    module.lambda_autoscale.name
  ]
}

resource "aws_lambda_permission" "autoscale_invoke" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_autoscale.name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.autoscale.arn

  depends_on = [
    module.lambda_autoscale.name
  ]
}
