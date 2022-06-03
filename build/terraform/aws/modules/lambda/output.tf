output "role" {
  value = aws_iam_role.role.name
}

output "arn" {
  value = aws_lambda_function.lambda_function.arn
}

output "name" {
  value = aws_lambda_function.lambda_function.function_name
}

output "secret_arn" {
  value = length(aws_secretsmanager_secret.secret) > 0 ? aws_secretsmanager_secret.secret[0].arn : null
}
