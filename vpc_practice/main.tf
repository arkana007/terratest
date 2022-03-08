provider "aws" {
  region = var.aws_region  
}
resource "aws_vpc" "vpc1" {
  cidr_block = var.vpc_cidr
  tags = {
      Name = var.vpc_name
  }
}


