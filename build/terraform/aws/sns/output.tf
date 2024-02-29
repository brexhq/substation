output "arn" {
  value       = aws_sns_topic.topic.arn
  description = "The ARN of the SNS topic."
}

output "id" {
  value       = aws_sns_topic.topic.id
  description = "The ID of the SNS topic."
}
