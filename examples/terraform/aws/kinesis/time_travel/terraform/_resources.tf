module "appconfig" {
  source = "../../../../../../build/terraform/aws/appconfig"

  config = {
    name         = "substation"
    environments = [{ name = "example" }]
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

# Kinesis Data Stream that stores data sent from pipeline sources.
module "kinesis" {
  source = "../../../../../../build/terraform/aws/kinesis_data_stream"

  config = {
    name              = "substation"
    autoscaling_topic = aws_sns_topic.autoscaling_topic.arn
    shards            = 1
  }

  access = [
    # Autoscales the stream.
    module.lambda_autoscaling.role.name,
    # Consumes data from the stream.
    module.lambda_enrichment.role.name,
    module.lambda_transform.role.name,
  ]
}

module "dynamodb" {
  source = "../../../../../../build/terraform/aws/dynamodb"

  config = {
    name      = "substation"
    hash_key  = "PK"
    range_key = "SK"
    ttl       = "TTL"

    attributes = [
      {
        name = "PK"
        type = "S"
      },
      {
        name = "SK"
        type = "S"
      },
    ]
  }

  access = [
    module.lambda_enrichment.role.name,
    module.lambda_transform.role.name,
  ]
}
