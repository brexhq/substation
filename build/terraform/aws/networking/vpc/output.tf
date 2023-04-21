output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "private_subnet_ids" {
  value = [
    for k in aws_subnet.private_subnet : k.id
  ]
}

output "default_security_group_id" {
  value = aws_vpc.vpc.default_security_group_id
}
