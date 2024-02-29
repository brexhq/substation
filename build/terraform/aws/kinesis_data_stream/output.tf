output "arn" {
  value       = aws_kinesis_stream.stream.arn
  description = "The ARN of the Kinesis Stream."
}

output "name" {
  value       = aws_kinesis_stream.stream.name
  description = "The name of the Kinesis Stream."
}
