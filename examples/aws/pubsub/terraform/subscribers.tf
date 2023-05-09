################################################
# Substation nodes that are subscribers to SNS
################################################

module "subscriber_x" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "subscriber_x"
  description   = "Substation node that is subscribed to SNS"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest3"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/subscriber_x"
    "SUBSTATION_HANDLER" : "AWS_SNS"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.repository_url,
  ]
}

resource "aws_sns_topic_subscription" "subscriber_x" {
  topic_arn = module.sns.arn
  protocol  = "lambda"
  endpoint  = module.subscriber_x.arn
}

resource "aws_lambda_permission" "subscriber_x" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.subscriber_x.name
  principal     = "sns.amazonaws.com"
  source_arn    = module.sns.arn

  depends_on = [
    module.subscriber_x.name
  ]
}

module "subscriber_y" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "subscriber_y"
  description   = "Substation node that is subscribed to SNS"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest3"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/subscriber_y"
    "SUBSTATION_HANDLER" : "AWS_SNS"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.repository_url,
  ]
}

resource "aws_sns_topic_subscription" "subscriber_y" {
  topic_arn = module.sns.arn
  protocol  = "lambda"
  endpoint  = module.subscriber_y.arn
}

resource "aws_lambda_permission" "subscriber_y" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.subscriber_y.name
  principal     = "sns.amazonaws.com"
  source_arn    = module.sns.arn

  depends_on = [
    module.subscriber_y.name
  ]
}


module "subscriber_z" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "subscriber_z"
  description   = "Substation node that is subscribed to SNS"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest3"
  architectures = ["arm64"]

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/subscriber_z"
    "SUBSTATION_HANDLER" : "AWS_SNS"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.repository_url,
  ]
}

resource "aws_sns_topic_subscription" "subscriber_z" {
  topic_arn = module.sns.arn
  protocol  = "lambda"
  endpoint  = module.subscriber_z.arn
}

resource "aws_lambda_permission" "subscriber_z" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = module.subscriber_z.name
  principal     = "sns.amazonaws.com"
  source_arn    = module.sns.arn

  depends_on = [
    module.subscriber_z.name
  ]
}
