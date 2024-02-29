module "lambda_node" {
  source    = "../../../../../../../build/terraform/aws/lambda"
  appconfig = module.appconfig

  config = {
    name        = "node"
    description = "Substation node that reads and writes data to S3."
    image_uri   = "${module.ecr.url}:latest"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_HANDLER" : "AWS_S3"
      "SUBSTATION_DEBUG" : true
    }
  }

  depends_on = [
    module.appconfig.name,
    module.ecr.url,
  ]
}

resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_node.name
  principal     = "s3.amazonaws.com"
  source_arn    = module.s3.arn
}
