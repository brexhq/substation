output "arn" {
  value = aws_kms_key.key.arn
}

output "id" {
  value = aws_kms_key.key.key_id
}
