output "arn" {
  value       = aws_appconfig_application.app.arn
  description = "The ARN of the AppConfig application."
}

output "id" {
  value       = aws_appconfig_application.app.id
  description = "The ID of the AppConfig application."
}

output "lambda" {
  value       = var.lambda
  description = "The validator Lambda function passed to the AppConfig application."
}
