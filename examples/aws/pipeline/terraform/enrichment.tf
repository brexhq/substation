################################################
# Lambda
# provides data enrichment as a microservice
################################################

module "lambda_enrichment" {
  source        = "../../../../build/terraform/aws/lambda"
  kms = module.kms
  appconfig = aws_appconfig_application.substation

  config = {
    name = "substation_lambda_enrichment"
    description = "Provides a microservice interface to Substation"
    image_uri = "${module.ecr_substation.url}:latest"
    architectures = ["arm64"]
    # Microservices require less memory and lower timeouts.
    memory = 128
    timeout     = 10
    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/prod/configurations/substation_lambda_enrichment"
      "SUBSTATION_HANDLER" : "AWS_LAMBDA_SYNC"
      "SUBSTATION_DEBUG" : true
    }
  }

  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.vpc,
  ]
}
