output "arn" {
  value       = aws_secretsmanager_secret.secret.arn
  description = "The ARN of the secret."
}
