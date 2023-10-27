resource "aws_vpc" "vpc" {
  cidr_block = var.config.cidr_block
  tags       = var.tags
}

# Public subnet resources
resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_subnet" "public_subnet" {
  vpc_id            = aws_vpc.vpc.id
  cidr_block        = keys(var.config.public_subnet)[0]
  availability_zone = values(var.config.public_subnet)[0]
  tags              = var.tags
}

resource "aws_route_table" "public_route" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.internet_gateway.id
  }

  tags = var.tags
}

# Private subnet resources
resource "aws_eip" "eip" {
  for_each = var.config.private_subnets

  domain = "vpc"
}

resource "aws_nat_gateway" "nat_gateway" {
  for_each = var.config.private_subnets

  allocation_id = aws_eip.eip[each.key].id
  subnet_id     = aws_subnet.public_subnet.id

  depends_on = [aws_internet_gateway.internet_gateway]
}

resource "aws_route_table_association" "public_subnet_association" {
  subnet_id      = aws_subnet.public_subnet.id
  route_table_id = aws_route_table.public_route.id
}

resource "aws_subnet" "private_subnet" {
  for_each = var.config.private_subnets

  vpc_id            = aws_vpc.vpc.id
  cidr_block        = each.key
  availability_zone = each.value
  tags              = var.tags
}

resource "aws_route_table" "private_route" {
  for_each = var.config.private_subnets

  vpc_id = aws_vpc.vpc.id
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.nat_gateway[each.key].id
  }

  tags = var.tags
}

resource "aws_route_table_association" "private_subnet_association" {
  for_each = var.config.private_subnets

  subnet_id      = aws_subnet.private_subnet[each.key].id
  route_table_id = aws_route_table.private_route[each.key].id
}
