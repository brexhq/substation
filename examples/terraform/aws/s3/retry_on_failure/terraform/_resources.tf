module "appconfig" {
  source = "../../../../../../build/terraform/aws/appconfig"

  config = {
    name         = "substation"
    environments = [{ name = "example" }]
  }
}

module "ecr" {
  source = "../../../../../../build/terraform/aws/ecr"

  config = {
    name         = "substation"
    force_delete = true
  }
}

resource "random_uuid" "s3" {}

# Monolithic S3 bucket used to store all data.
module "s3" {
  source = "../../../../../../build/terraform/aws/s3"

  config = {
    # Bucket name is randomized to avoid collisions.
    name = "${random_uuid.s3.result}-substation"
  }

  access = [
    # Reads objects from the bucket.
    module.lambda_node.role.name,
  ]
}

module "sqs" {
  source = "../../../../../../build/terraform/aws/sqs"

  config = {
    name = "substation_retries"

    # Delay for 30 seconds to allow the pipeline to recover.
    delay = 30

    # Timeout should be at least 6x the timeout of any consuming Lambda functions, plus the batch window.
    # Refer to the Lambda documentation for more information: 
    # https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html.
    timeout = 60
  }

  access = [
    # Sends messages to the queue.
    module.lambda_retrier.role.name,
    # Receives messages from the queue.
    module.lambda_node.role.name,
  ]
}
