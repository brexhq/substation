# Used for testing Substation configurations via AWS AppConfig Validation or with manual invocations.

# first runs of this Terraform will fail due to an empty ECR image
module "lambda_validator" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "substation_validator"
  description   = "Validates Substation configurations via AWS AppConfig Validation or manual invocations."
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_validation.repository_url}:latest"
  architectures = ["arm64"]

  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_validation.repository_url,
  ]
}

resource "aws_lambda_permission" "validator_invoke" {
  statement_id  = "AllowExecutionFromAppConfig"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_validator.name
  principal     = "appconfig.amazonaws.com"

  depends_on = [
    module.lambda_validator.name
  ]
}
