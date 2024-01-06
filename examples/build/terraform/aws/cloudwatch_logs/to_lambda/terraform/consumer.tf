data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

module "lambda_consumer" {
  source = "../../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "consumer"
    description = "Substation node that is invoked by CloudWatch"
    image_uri   = "${module.ecr_substation.url}:latest"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/consumer"
      "SUBSTATION_HANDLER" : "AWS_LAMBDA"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.url,
  ]
}

# Allows any CloudWatch log group to send logs to the Lambda function in the current AWS account and region.
# Repeat this for each region that sends logs to the Lambda function.
resource "aws_lambda_permission" "consumer" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_consumer.name
  principal     = "logs.amazonaws.com"
  source_arn    = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:*"
}

# CloudWatch Logs subscription filter that sends logs to the Lambda function.
module "cw_subscription" {
  source = "../../../../../../../build/terraform/aws/cloudwatch/subscription"

  config = {
    name            = "substation"
    destination_arn = module.lambda_consumer.arn
    log_groups = [
      # This example causes recursion. Add other log groups for resources in the account and region.
      "/aws/lambda/test",
    ]
  }
}
