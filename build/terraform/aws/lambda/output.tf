output "role" {
  value       = aws_iam_role.role
  description = "The IAM role used by the Lambda function."
}

output "arn" {
  value       = aws_lambda_function.lambda_function.arn
  description = "The ARN of the Lambda function."
}

output "name" {
  value       = aws_lambda_function.lambda_function.function_name
  description = "The name of the Lambda function."
}
