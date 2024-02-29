output "arn" {
  value       = aws_kms_key.key.arn
  description = "The ARN of the KMS key."
}

output "id" {
  value       = aws_kms_key.key.key_id
  description = "The ID of the KMS key."
}
