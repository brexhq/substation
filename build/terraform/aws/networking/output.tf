output "vpc_id" {
  value = aws_vpc.substation_vpc.id
}

output "private_subnet_1" {
    value = aws_subnet.private_subnet_1.id
}

output "private_subnet_2" {
    value = aws_subnet.private_subnet_2.id
}

output "private_subnet_3" {
    value = aws_subnet.private_subnet_3.id
}