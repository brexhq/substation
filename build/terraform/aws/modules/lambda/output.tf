output "role" {
  value = aws_iam_role.role.name
}

output "arn" {
  value = aws_lambda_function.lambda_function.arn
}

output "name" {
  value = aws_lambda_function.lambda_function.function_name
}
