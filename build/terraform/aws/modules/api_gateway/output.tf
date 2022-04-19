output "execution_arn" {
  value = aws_api_gateway_rest_api.api.execution_arn
}

output "id" {
  value = aws_api_gateway_rest_api.api.id
}

output "resource_id" {
  value = aws_api_gateway_method.method.resource_id
}

output "http_method" {
  value = aws_api_gateway_method.method.http_method
}
