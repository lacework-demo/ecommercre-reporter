provider "aws" {
  region = "us-east-1"
}

resource "random_string" "random" {
  length  = 10
  special = false
  upper   = false
}

locals {
  bucket_name = "sko-bucket-${random_string.random.result}"
}

resource "aws_s3_bucket" "bucket" {
  bucket = local.bucket_name

  tags = {
    Name        = local.bucket_name
    Environment = "sko"
  }
  force_destroy = true
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

output "bucket" {
  value = aws_s3_bucket.bucket.id
}
