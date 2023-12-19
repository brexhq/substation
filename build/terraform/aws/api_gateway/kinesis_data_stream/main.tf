resource "random_uuid" "id" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_api_gateway_rest_api" "api" {
  name = var.config.name
  tags = var.tags
}

resource "aws_api_gateway_method" "method" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_rest_api.api.root_resource_id
  http_method   = "POST"
  authorization = "NONE"
}

data "aws_iam_policy_document" "service_policy_document" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = [
        "apigateway.amazonaws.com"
      ]
    }
  }
}

resource "aws_iam_role" "role" {
  name               = "substation-api-gateway-${resource.random_uuid.id.id}"
  assume_role_policy = data.aws_iam_policy_document.service_policy_document.json
  tags               = var.tags
}

resource "aws_api_gateway_integration" "gateway_integration" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_method.method.resource_id
  http_method             = aws_api_gateway_method.method.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  timeout_milliseconds    = var.config.timeout
  uri = format(
    "arn:%s:apigateway:%s:kinesis:action/PutRecord",
    data.aws_partition.current.partition,
    data.aws_region.current.name
  )
  credentials = aws_iam_role.role.arn
  request_parameters = {
    "integration.request.header.Content-Type" = "'application/x-amz-json-1.1'"
  }
  request_templates = {
    "application/json" = <<EOF
    {
        "StreamName": "${var.kinesis_data_stream.name}",
        "Data": "$util.base64Encode($input.body)",
        "PartitionKey": "$context.requestId"
    }
    EOF
  }
}

resource "aws_api_gateway_method_response" "response_200" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_method.method.resource_id
  http_method = aws_api_gateway_method.method.http_method
  status_code = "200"
  response_models = {
    "application/json" = "Empty"
  }
  response_parameters = {}
}

resource "aws_api_gateway_integration_response" "integration_response" {
  rest_api_id         = aws_api_gateway_rest_api.api.id
  resource_id         = aws_api_gateway_method.method.resource_id
  http_method         = aws_api_gateway_method.method.http_method
  status_code         = aws_api_gateway_method_response.response_200.status_code
  selection_pattern   = "200"
  response_parameters = {}

  depends_on = [
    aws_api_gateway_integration.gateway_integration,
  ]
}

resource "aws_api_gateway_deployment" "deployment" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = var.config.name

  depends_on = [
    aws_api_gateway_integration.gateway_integration,
    aws_api_gateway_integration_response.integration_response,
  ]
}
