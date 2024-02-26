output "arn" {
  value       = aws_sqs_queue.queue.arn
  description = "The ARN of the SQS queue."
}

output "id" {
  value       = aws_sqs_queue.queue.id
  description = "The ID of the SQS queue."
}
