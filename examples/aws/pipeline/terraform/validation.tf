# Used for testing Substation configurations via AWS AppConfig Validation or with manual invocations.

module "lambda_validator" {
  source        = "../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "substation_validator"
    description = "Validates configurations"
    image_uri = "${module.ecr_validation.url}:latest"
    architectures = ["arm64"]
  }

  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.url,
    module.vpc,
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
