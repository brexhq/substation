module "subscriber_x" {
  source = "../../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "subscriber_x"
    description = "Substation node that subscribes to SNS"
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/subscriber_x"
      "SUBSTATION_HANDLER" : "AWS_SNS"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
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
  source = "../../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "subscriber_y"
    description = "Substation node that subscribes to SNS"
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/subscriber_y"
      "SUBSTATION_HANDLER" : "AWS_SNS"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
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
  source = "../../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig  

  config = {
    name        = "subscriber_z"
    description = "Substation node that subscribes to SNS"
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/subscriber_z"
      "SUBSTATION_HANDLER" : "AWS_SNS"
      "SUBSTATION_DEBUG" : true
    }
  }


  depends_on = [
    module.appconfig.name,
    module.ecr.url,
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
