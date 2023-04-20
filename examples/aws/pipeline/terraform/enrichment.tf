################################################
# Lambda
# provides data enrichment as a microservice
################################################

module "lambda_enrichment" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "substation_lambda_enrichment"
  description   = "Substation Lambda that is triggered synchronously and provides enrichment as a microservice"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]
  # use lower memory and timeouts for microservice deployments
  memory_size = 128
  timeout     = 10

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_lambda_enrichment"
    "SUBSTATION_HANDLER" : "AWS_LAMBDA_SYNC"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  vpc_config = {
    subnet_ids = [
      module.network.private_subnet_id,
      module.network.public_subnet_id,
    ]
    security_group_ids = [module.network.public_egress_security_group_id]
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_autoscaling.repository_url,
    module.network,
  ]
}
