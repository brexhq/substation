resource "random_uuid" "id" {}

# var.map[*] is a convenience function for handling empty maps.
locals {
  env = var.config.env[*]
}

resource "aws_lambda_function" "lambda_function" {
  function_name = var.config.name
  description   = var.config.description

  # Runtime settings.
  role        = aws_iam_role.role.arn
  kms_key_arn = var.kms != null ? var.kms.arn : null
  timeout     = var.config.timeout
  memory_size = var.config.memory

  # Architecture settings.
  package_type  = "Image" # Only container images are supported.
  image_uri     = var.config.image_uri
  architectures = var.config.image_arm ? ["arm64"] : ["x86_64"]


  # Network settings.
  vpc_config {
    subnet_ids         = var.config.vpc_config.subnet_ids
    security_group_ids = var.config.vpc_config.security_group_ids
  }

  # Tracing settings.
  tracing_config {
    mode = "Active"
  }

  # Environment settings.
  # Required for avoiding errors due to missing environment variables.
  dynamic "environment" {
    for_each = local.env
    content {
      variables = environment.value
    }
  }

  tags = var.tags
}

resource "aws_iam_role" "role" {
  name               = "substation-lambda-${resource.random_uuid.id.id}"
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

resource "aws_appconfig_configuration_profile" "config" {
  count = var.appconfig != null ? 1 : 0

  application_id = var.appconfig.id
  description    = "Configuration profile for the ${var.config.name} Lambda"
  name           = var.config.name
  location_uri   = "hosted"

  dynamic "validator" {
    for_each = var.appconfig.lambda != null ? [1] : []
    content {
      type    = "LAMBDA"
      content = var.appconfig.lambda.arn
    }
  }

  tags = var.tags
}

################################################
# Default Policies
################################################

resource "aws_iam_role_policy_attachment" "lambda_basic_execution_role" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_vpc_access_execution_role" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_iam_role_policy_attachment" "xray_write_only_access" {
  role       = aws_iam_role.role.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXrayWriteOnlyAccess"
}

################################################
# Custom Policies
################################################

resource "aws_iam_role_policy_attachment" "custom_policy_attachment" {
  role       = aws_iam_role.role.name
  policy_arn = aws_iam_policy.custom_policy.arn
}

resource "aws_iam_policy" "custom_policy" {
  name        = "substation-lambda-${resource.random_uuid.id.id}"
  description = "Policy for the ${var.config.name} Lambda."
  policy      = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
  // This is a no-op statement that is used for creating a default policy.
  statement {
    effect = "Allow"
    actions = [
      "none:Substation",
    ]

    resources = [
      "*"
    ]
  }

  dynamic "statement" {
    for_each = var.appconfig != null ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "appconfig:GetConfiguration",
        "appconfig:GetLatestConfiguration",
        "appconfig:StartConfigurationSession",
      ]

      resources = [
        "${var.appconfig.arn}/*"
      ]
    }
  }

  dynamic "statement" {
    for_each = var.kms != null ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ]

      resources = [
        var.kms.arn,
      ]
    }
  }

  // Add additional statements provided as a variable.
  dynamic "statement" {
    for_each = var.config.iam_statements
    content {
      effect    = "Allow"
      actions   = statement.value.actions
      resources = statement.value.resources
    }
  }
}

################################################
# Access Policies
################################################

resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "substation-lambda-access-${resource.random_uuid.id.id}"
  description = "Policy that grants access to the Substation ${var.config.name} Lambda."
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"
    actions = [
      "lambda:GetFunctionConfiguration",
      "lambda:InvokeAsync",
      "lambda:InvokeFunction",
    ]

    resources = [
      aws_lambda_function.lambda_function.arn,
      # This is required for data transformation support in Kinesis Data Firehose.
      "${aws_lambda_function.lambda_function.arn}:*", # Allow access to all versions.
    ]
  }

  dynamic "statement" {
    for_each = var.kms != null ? [1] : []

    content {
      effect = "Allow"
      actions = [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ]

      resources = [
        var.kms.arn,
      ]
    }
  }
}
