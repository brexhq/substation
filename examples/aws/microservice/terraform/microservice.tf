################################################
# Lambda
################################################

module "microservice" {
  source        = "../../../../build/terraform/aws/lambda"
  function_name = "substation_microservice"
  description   = "Substation Lambda that is triggered synchronously as a microservice"
  appconfig_id  = aws_appconfig_application.substation.id
  kms_arn       = module.kms_substation.arn
  image_uri     = "${module.ecr_substation.repository_url}:latest"
  architectures = ["arm64"]
  # use lower memory and timeouts for microservice deployments
  memory_size   = 128
  timeout       = 10

  env = {
    "AWS_MAX_ATTEMPTS" : 10
    "AWS_APPCONFIG_EXTENSION_PREFETCH_LIST" : "/applications/substation/environments/prod/configurations/substation_microservice"
    "SUBSTATION_HANDLER" : "AWS_LAMBDA_SYNC"
    "SUBSTATION_DEBUG" : 1
    "SUBSTATION_METRICS" : "AWS_CLOUDWATCH_EMBEDDED_METRICS"
  }
  tags = {
    owner = "example"
  }

  depends_on = [
    aws_appconfig_application.substation,
    module.ecr_substation.repository_url,
  ]
}

################################################
# appconfig permissions
# all Lambda must have this policy
################################################

module "iam_appconfig_read" {
  source    = "../../../../build/terraform/aws/iam"
  resources = ["${aws_appconfig_application.substation.arn}/*"]
}

module "iam_appconfig_read_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_appconfig_read"
  policy = module.iam_appconfig_read.appconfig_read_policy
  roles = [
    module.microservice.role,
  ]
}

################################################
# KMS read permissions
# all Lambda must have this policy
################################################

module "iam_kms_read" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.kms_substation.arn,
  ]
}

module "iam_kms_read_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_kms_read"
  policy = module.iam_kms_read.kms_read_policy
  roles = [
    module.microservice.role,
  ]
}

################################################
# KMS write permissions
# all Lambda must have this policy
################################################

module "iam_kms_write" {
  source = "../../../../build/terraform/aws/iam"
  resources = [
    module.kms_substation.arn,
  ]
}

module "iam_kms_write_attachment" {
  source = "../../../../build/terraform/aws/iam_attachment"
  id     = "substation_kms_write"
  policy = module.iam_kms_write.kms_write_policy
  roles = [
    module.microservice.role,
  ]
}
