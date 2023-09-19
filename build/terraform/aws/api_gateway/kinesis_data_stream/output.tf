output "role" {
  value = aws_iam_role.role
}

output "url" {
  value = aws_api_gateway_deployment.deployment.invoke_url
}
