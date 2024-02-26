output "role" {
  value       = aws_iam_role.role
  description = "The IAM role used by the API Gateway."
}

output "url" {
  value       = aws_api_gateway_deployment.deployment.invoke_url
  description = "The URL of the API Gateway."
}
