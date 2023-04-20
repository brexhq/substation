# Main substation VPC
resource "aws_vpc" "vpc" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = var.instance_tenancy
  tags             = var.tags
}

# Create private subnets
resource "aws_subnet" "private_subnet" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = var.subnet_cidr
}

# Create egress only IGW
resource "aws_egress_only_internet_gateway" "substation_egress" {
  vpc_id = aws_vpc.vpc.id
}

# Create routes
resource "aws_route_table" "private_route" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    # Default route to the internet, no IGW attached
  }
  route {
    cidr_block             = "0.0.0.0/0"
    egress_only_gateway_id = aws_egress_only_internet_gateway.substation_egress.id
  }
}

# Private subnet route associations
resource "aws_route_table_association" "private_subnet_association" {
  subnet_id      = aws_subnet.private_subnet.id
  route_table_id = aws_route_table.private_route.id
}

# Default TLS security group for substation VPC

resource "aws_security_group" "allow_substation_tls" {
  name        = "allow_tls"
  description = "Allow TLS inbound traffic"
  vpc_id      = aws_vpc.vpc.id

  ingress {
    description = "TLS from VPC"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "allow_tls"
  }
}