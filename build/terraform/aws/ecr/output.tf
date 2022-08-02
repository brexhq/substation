output "repository_url" {
  value       = aws_ecr_repository.repository.repository_url
  description = "The URL for the ECR repository"
}
