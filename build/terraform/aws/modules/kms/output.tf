output "arn" {
  value = aws_kms_key.key.arn
}

output "key_id" {
  value = aws_kms_key.key.key_id
}
