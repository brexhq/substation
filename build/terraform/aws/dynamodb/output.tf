output "arn" {
  value = aws_dynamodb_table.table.arn
}

output "stream_arn" {
  value = aws_dynamodb_table.table.stream_arn
}
