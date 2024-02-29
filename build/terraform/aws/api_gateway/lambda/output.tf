output "url" {
  value       = aws_api_gateway_deployment.deployment.invoke_url
  description = "The URL of the API Gateway."
}

output "name" {
  value       = aws_api_gateway_rest_api.api.name
  description = "The name of the API Gateway REST API."
}
