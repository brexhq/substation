output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "private_subnet_id" {
  value = aws_subnet.private_subnet.id
}

output "public_subnet_id" {
  value = aws_subnet.public_subnet.id
}

output "public_egress_security_group_id" {
  value = aws_security_group.public_egress.id
}
