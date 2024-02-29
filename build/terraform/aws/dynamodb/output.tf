output "arn" {
  value       = aws_dynamodb_table.table.arn
  description = "The ARN of the DynamoDB table."
}

output "stream_arn" {
  value       = aws_dynamodb_table.table.stream_arn
  description = "The ARN of the DynamoDB table stream."
}
