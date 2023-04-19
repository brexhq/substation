# var.map[*] is a convenience function for handling empty maps
locals {
  env = var.env[*]
}

module "vpc" {
  source = "../networking"
}

resource "aws_lambda_function" "lambda_function" {
  function_name = var.function_name
  description   = var.description
  image_uri     = var.image_uri
  package_type  = "Image"
  architectures = var.architectures
  role          = aws_iam_role.role.arn
  timeout       = var.timeout
  memory_size   = var.memory_size

  vpc_config {
    subnet_ids         = var.use_substation_vpc ? [module.vpc.private_subnet.id] : var.subnet_ids
    security_group_ids = var.use_substation_vpc ? [module.vpc.allow_substation_tls.id] : var.security_group_ids
  }

  tracing_config {
    mode = "Active"
  }

  # required for avoiding errors due to missing environment variables
  dynamic "environment" {
    for_each = local.env
    content {
      variables = environment.value
    }
  }

  kms_key_arn = var.kms_arn
  tags        = var.tags
}

resource "aws_iam_role" "role" {
  name               = "${var.function_name}_role"
  assume_role_policy = data.aws_iam_policy_document.service_policy_document.json

  tags = var.tags
}

data "aws_iam_policy_document" "service_policy_document" {
  statement {
    sid    = "AssumeRole"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role_policy_attachment" "xray_write_only_access" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
}

# permissions for writing logs to CloudWatch
resource "aws_iam_role_policy_attachment" "lambda_basic_execution_role" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_appconfig_configuration_profile" "config" {
  application_id = var.appconfig_id
  description    = "Configuration profile for the ${var.function_name} Lambda"
  name           = var.function_name
  location_uri   = "hosted"

  tags = var.tags
}

# optional secret creation
resource "aws_secretsmanager_secret" "secret" {
  count      = var.secret ? 1 : 0
  name       = var.function_name
  kms_key_id = var.kms_arn
  tags       = var.tags
}
