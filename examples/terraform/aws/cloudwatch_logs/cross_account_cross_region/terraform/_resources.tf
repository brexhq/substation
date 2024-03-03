data "aws_caller_identity" "caller" {}

module "appconfig" {
  source = "../../../../../../build/terraform/aws/appconfig"

  config = {
    name = "substation"
    environments = [{
      name = "example"
    }]
  }
}

# Repository for the core Substation application.
module "ecr" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}

# Repository for the autoscaling application.
module "ecr_autoscale" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "autoscale"
    force_delete = true
  }
}

# SNS topic for Kinesis Data Stream autoscaling alarms.
resource "aws_sns_topic" "autoscaling_topic" {
  name = "autoscale"
}

# Kinesis Data Stream that is used as the destination for CloudWatch Logs.
module "kds" {
  source = "../../../../../../build/terraform/aws/kinesis_data_stream"

  config = {
    name              = "substation"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
  }

  access = [
    # Autoscales the stream.
    module.lambda_autoscaling.role.name,
    # Reads data from the stream.
    module.lambda_consumer.role.name,
    # Writes data to the stream.
    module.cw_destination_use1.role.name,
    module.cw_destination_usw2.role.name,
  ]
}

# CloudWatch Logs destination that sends logs to the Kinesis Data Stream from us-east-1.
module "cw_destination_use1" {
  source = "../../../../../../build/terraform/aws/cloudwatch/destination"

  config = {
    name            = "substation"
    destination_arn = module.kds.arn

    # By default, any CloudWatch log in the current AWS account can send logs to this destination.
    # Add additional AWS account IDs to allow them to send logs to the destination.
    account_ids = []
  }
}

module "cw_subscription_use1" {
  source = "../../../../../../build/terraform/aws/cloudwatch/subscription"

  config = {
    name            = "substation"
    destination_arn = module.cw_destination_use1.arn
    log_groups = [
      # This example causes recursion. Add other log groups for resources in us-east-1.
      # "/aws/lambda/consumer",
    ]
  }
}

# CloudWatch Logs destination that sends logs to the Kinesis Data Stream from us-west-2.
# To add support for more regions, copy this module and change the provider.
module "cw_destination_usw2" {
  source = "../../../../../../build/terraform/aws/cloudwatch/destination"
  providers = {
    aws = aws.usw2
  }

  kms = module.kms
  config = {
    name            = "substation"
    destination_arn = module.kds.arn

    # By default, any CloudWatch log in the current AWS account can send logs to this destination.
    # Add additional AWS account IDs to allow them to send logs to the destination.
    account_ids = []
  }
}
