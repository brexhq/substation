output "vpc_id" {
  value       = aws_vpc.vpc.id
  description = "The ID of the VPC."
}

output "public_subnet_id" {
  value       = aws_subnet.public_subnet.id
  description = "The ID of the public subnet."
}

output "private_subnet_ids" {
  value = [
    for k in aws_subnet.private_subnet : k.id
  ]
  description = "The IDs of the private subnets."
}

output "default_security_group_id" {
  value       = aws_vpc.vpc.default_security_group_id
  description = "The ID of the default security group."
}
