resource "aws_iam_role" "ecs_role" {
  name = "ecs_role"

  assume_role_policy = jsonencode({
    "Version": "2008-10-17",
    "Statement": [
      {
          "Sid": "",
          "Effect": "Allow",
          "Principal": {
              "Service": "ecs-tasks.amazonaws.com"
          },
          "Action": "sts:AssumeRole"
      }
    ]
  })

  inline_policy {
    name = "AmazonEC2ContainerRegistryFullAccess"

    policy = jsonencode({
      "Version": "2012-10-17",
      "Statement": [
          {
              "Effect": "Allow",
              "Action": [
                  "logs:*",
                  "ecr:*",
                  "cloudtrail:LookupEvents"
              ],
              "Resource": "*"
          },
          {
              "Effect": "Allow",
              "Action": [
                  "iam:CreateServiceLinkedRole"
              ],
              "Resource": "*",
              "Condition": {
                  "StringEquals": {
                      "iam:AWSServiceName": [
                          "replication.ecr.amazonaws.com"
                      ]
                  }
              }
          },
      ]
    })
  }
}
