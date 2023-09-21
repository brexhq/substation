output "url" {
  value = aws_api_gateway_deployment.deployment.invoke_url
}

output "name" {
  value = aws_api_gateway_rest_api.api.name
}
