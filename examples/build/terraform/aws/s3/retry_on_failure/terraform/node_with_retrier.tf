module "lambda_node" {
  source = "../../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "node"
    description = "Substation node that reads data from S3. The node will retry forever if it fails."
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_HANDLER" : "AWS_S3"
      "SUBSTATION_DEBUG" : true
    }
  }

  # The retrier Lambda must be able to invoke this 
  # Lambda function to retry failed S3 events.
  access = [
    module.lambda_retrier.role.name,
  ]

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr.url,
  ]
}

resource "aws_lambda_permission" "node" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_node.name
  principal     = "s3.amazonaws.com"
  source_arn    = module.s3.arn
}

resource "aws_s3_bucket_notification" "node" {
  bucket = module.s3.id

  lambda_function {
    lambda_function_arn = module.lambda_node.arn
    events              = ["s3:ObjectCreated:*"]
  }

  depends_on = [
    aws_lambda_permission.node
  ]
}

# Configures the Lambda function to send failed events to the SQS queue.
resource "aws_lambda_function_event_invoke_config" "node" {
  function_name = module.lambda_node.name

  # This example disables the built-in retry mechanism.
  maximum_retry_attempts = 0

  destination_config {
    on_failure {
      destination = module.sqs_queue.arn
    }
  }
}

module "lambda_retrier" {
  source = "../../../../../../../build/terraform/aws/lambda"
  # These are always required for all Lambda.
  kms       = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name        = "retrier"
    description = "Substation node that receives events from the retry queue and invokes the original Lambda function."
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true

    # This value must match the timeout of the SQS queue.
    timeout = 30

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/retrier"
      "SUBSTATION_HANDLER" : "AWS_SQS"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr.url,
  ]
}

resource "aws_lambda_event_source_mapping" "retrier" {
  event_source_arn                   = module.sqs_queue.arn
  function_name                      = module.lambda_retrier.arn
  maximum_batching_window_in_seconds = 30
  batch_size                         = 10
}
