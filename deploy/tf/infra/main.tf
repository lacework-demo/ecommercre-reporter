variable "vpc_id" {
  type = string
}

resource "aws_security_group" "reporter-all-intra-traffic" {
  name   = "ecomm-reporter-internal-traffic"
  vpc_id = var.vpc_id

  ingress {
    protocol  = "ALL"
    self      = true
    from_port = 0
    to_port   = 0
  }
}

resource "random_string" "random" {
  length  = 10
  special = false
  upper   = false
}

locals {
  name_suffix = random_string.random.result
}


# EC2 IAM role setup
resource "aws_iam_policy" "reporter_ec2_policy" { // ec2 policy
  name = "reporter_ec2_policy_${local.name_suffix}"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : "*",
        "Resource" : "*"
      }
    ]
  })
}

resource "aws_iam_role" "reporter_ec2_role" { // assume role
  name = "reporter_ec2_role_${local.name_suffix}"
  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "ec2.amazonaws.com"
        },
        "Action" : "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_policy_attachment" "reporter_ec2_role_attachment" {
  name       = "reporter_ec2_role_attach_${local.name_suffix}"
  roles      = [aws_iam_role.reporter_ec2_role.id]
  policy_arn = aws_iam_policy.reporter_ec2_policy.id
}

resource "aws_iam_instance_profile" "reporter_ec2_instance_profile" {
  name = "reporter_ec2_instance_profile_${local.name_suffix}"
  role = aws_iam_role.reporter_ec2_role.id
}

output "security_group" {
  value = aws_security_group.reporter-all-intra-traffic.id
}

output "instance_profile" {
  value = aws_iam_instance_profile.reporter_ec2_instance_profile.id
}

output "name_suffix" {
  value = local.name_suffix
}

