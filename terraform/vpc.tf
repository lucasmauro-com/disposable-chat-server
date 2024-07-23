resource "aws_vpc" "disposable_chat" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_subnet" "us_east_1a" {
  vpc_id            = aws_vpc.disposable_chat.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}

resource "aws_subnet" "us_east_1b" {
  vpc_id            = aws_vpc.disposable_chat.id
  cidr_block        = "10.0.25.0/24"
  availability_zone = "us-east-1b"
}

resource "aws_subnet" "us_east_1c" {
  vpc_id            = aws_vpc.disposable_chat.id
  cidr_block        = "10.0.50.0/24"
  availability_zone = "us-east-1c"
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.disposable_chat.id
}

resource "aws_route_table" "routes" {
  vpc_id = aws_vpc.disposable_chat.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }
}

resource "aws_route_table_association" "subnet_1a" {
  subnet_id      = aws_subnet.us_east_1a.id
  route_table_id = aws_route_table.routes.id
}

resource "aws_route_table_association" "subnet_1b" {
  subnet_id      = aws_subnet.us_east_1b.id
  route_table_id = aws_route_table.routes.id
}

resource "aws_route_table_association" "subnet_1c" {
  subnet_id      = aws_subnet.us_east_1c.id
  route_table_id = aws_route_table.routes.id
}

resource "aws_elasticache_subnet_group" "disposable_chat" {
  name       = "disposable-chat-server-subnet-group"
  subnet_ids = [
    aws_subnet.us_east_1a.id,
    aws_subnet.us_east_1b.id,
    aws_subnet.us_east_1c.id,
  ]
}

resource "aws_security_group" "disposable_chat" {
  name   = "disposable-chat"
  vpc_id = aws_vpc.disposable_chat.id
}

resource "aws_vpc_security_group_ingress_rule" "allow_1a" {
  security_group_id = aws_security_group.disposable_chat.id
  cidr_ipv4         = aws_subnet.us_east_1a.cidr_block
  from_port         = -1
  to_port           = -1
  ip_protocol       = -1
}

resource "aws_vpc_security_group_ingress_rule" "allow_1b" {
  security_group_id = aws_security_group.disposable_chat.id
  cidr_ipv4         = aws_subnet.us_east_1b.cidr_block
  from_port         = -1
  to_port           = -1
  ip_protocol       = -1
}

resource "aws_vpc_security_group_ingress_rule" "allow_1c" {
  security_group_id = aws_security_group.disposable_chat.id
  cidr_ipv4         = aws_subnet.us_east_1c.cidr_block
  from_port         = -1
  to_port           = -1
  ip_protocol       = -1
}

resource "aws_vpc_security_group_ingress_rule" "allow_tcp_443" {
  security_group_id = aws_security_group.disposable_chat.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = "443"
  to_port           = "443"
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "allow_all" {
  cidr_ipv4 = "0.0.0.0/0"
  security_group_id = aws_security_group.disposable_chat.id
  ip_protocol       = -1
}

resource "aws_vpc_endpoint" "ecr_api" {
  vpc_id              = aws_vpc.disposable_chat.id
  service_name        = "com.amazonaws.us-east-1.ecr.api"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = [
    aws_subnet.us_east_1a.id,
    aws_subnet.us_east_1b.id,
    aws_subnet.us_east_1c.id,
  ]
  security_group_ids  = [aws_security_group.disposable_chat.id]
  private_dns_enabled = true
}

resource "aws_vpc_endpoint" "ecr_dkr" {
  vpc_id              = aws_vpc.disposable_chat.id
  service_name        = "com.amazonaws.us-east-1.ecr.dkr"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = [
    aws_subnet.us_east_1a.id,
    aws_subnet.us_east_1b.id,
    aws_subnet.us_east_1c.id,
  ]
  security_group_ids  = [aws_security_group.disposable_chat.id]
  private_dns_enabled = true
}