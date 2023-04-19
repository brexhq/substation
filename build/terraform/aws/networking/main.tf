# Main substation VPC
resource "aws_vpc" "substation_vpc" {
  cidr_block       = var.vpc_network_cidr
  instance_tenancy = var.instance_tenancy
  tags             = var.tags
}

# Create private subnets
resource "aws_subnet" "private_subnet_1" {
  vpc_id     = aws_vpc.substation_vpc.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "private_subnet_2" {
  vpc_id     = aws_vpc.substation_vpc.id
  cidr_block = "10.0.2.0/24"
}

resource "aws_subnet" "private_subnet_3" {
  vpc_id     = aws_vpc.substation_vpc.id
  cidr_block = "10.0.3.0/24"
}

# Create egress only IGW
resource "aws_egress_only_internet_gateway" "substation_egress" {
  vpc_id = aws_vpc.substation_vpc.id
}

# Create routes
resource "aws_route_table" "private_rt" {
  vpc_id = aws_vpc.substation_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    # Default route to the internet, no IGW attached
  }
  route {
    ipv6_cidr_block        = "::/0"
    egress_only_gateway_id = aws_egress_only_internet_gateway.substation_egress.id
  }
}

# Private subnet route associations
resource "aws_route_table_association" "private_subnet_1_association" {
  subnet_id      = aws_subnet.private_subnet_1.id
  route_table_id = aws_route_table.private_rt.id
}

resource "aws_route_table_association" "private_subnet_2_association" {
  subnet_id      = aws_subnet.private_subnet_2.id
  route_table_id = aws_route_table.private_rt.id
}

resource "aws_route_table_association" "private_subnet_3_association" {
  subnet_id      = aws_subnet.private_subnet_3.id
  route_table_id = aws_route_table.private_rt.id
}

# Default TLS security group for substation VPC

resource "aws_security_group" "allow_substation_tls" {
  name        = "allow_tls"
  description = "Allow TLS inbound traffic"
  vpc_id      = aws_vpc.substation_vpc.id

  ingress {
    description      = "TLS from VPC"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = [aws_vpc.substation_vpc.cidr_block]
    ipv6_cidr_blocks = [aws_vpc.substation_vpc.ipv6_cidr_block]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = "allow_tls"
  }
}