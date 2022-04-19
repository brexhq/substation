resource "aws_api_gateway_rest_api" "api" {
  name = var.name
  tags = var.tags
}

resource "aws_api_gateway_method" "method" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_rest_api.api.root_resource_id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "integration" {
  rest_api_id = module.example_gateway.id
  resource_id = module.example_gateway.resource_id
  http_method = module.example_gateway.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${data.aws_region.current.name}:lambda:path/2015-03-31/functions/${module.example_gateway_lambda.arn}/invocations"
}

resource "aws_api_gateway_deployment" "deployment" {
  depends_on = [
    aws_api_gateway_integration.integration,
  ]

  rest_api_id = module.example_gateway.id
  stage_name  = "substation"
}

resource "aws_lambda_permission" "gateway_permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.example_gateway_lambda.name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${module.example_gateway.execution_arn}/*/*"
}
