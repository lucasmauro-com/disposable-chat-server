data "aws_ecr_image" "server" {
  repository_name = "disposable-chat-server"
  most_recent       = true
}

resource "aws_ecs_cluster" "cluster" {
  name = "disposable-chat-server"

  setting {
    name  = "containerInsights"
    value = "disabled"
  }
}

resource "aws_ecs_task_definition" "task" {
  family                   = "disposable-chat-server"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 1024
  memory                   = 2048
  execution_role_arn       = aws_iam_role.ecs_role.arn
  container_definitions    = jsonencode([
    {
      name      = "disposable-chat-server"
      image     = data.aws_ecr_image.server.image_uri
      cpu       = 1024
      memory    = 2048
      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
        }
      ],
      environment: [
        {
          "name": "SERVER_PORT",
          "value": "80"
        },
        {
          "name": "ACCEPTED_ORIGIN",
          "value": "https://chat.lucasmauro.com"
        },
        {
          "name": "REDIS_ENDPOINT",
          "value": "${aws_elasticache_cluster.redis.cache_nodes[0].address}:${aws_elasticache_cluster.redis.cache_nodes[0].port}"
        },
        {
          "name": "DEBUG",
          "value": "true"
        }
      ],
      "logConfiguration": {
          "logDriver": "awslogs",
          "options": {
              "awslogs-group": "/ecs/disposable-chat-server",
              "awslogs-region": "us-east-1"
              "awslogs-stream-prefix": "disposable-chat",
              "awslogs-create-group": "true"
          }
      },
    }
  ])
}

resource "aws_ecs_service" "service" {
  name                 = "disposable-chat-server"
  cluster              = aws_ecs_cluster.cluster.id
  task_definition      = aws_ecs_task_definition.task.arn
  desired_count        = 2
  force_new_deployment = true

  enable_ecs_managed_tags = true
  wait_for_steady_state   = true

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    base              = 0
    weight            = 1
  }
  
  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  network_configuration {
    security_groups  = [aws_security_group.disposable_chat.id]
    subnets          = [aws_subnet.us_east_1a.id, aws_subnet.us_east_1b.id, aws_subnet.us_east_1c.id]
    assign_public_ip = true
  }
}

resource "aws_ecs_cluster_capacity_providers" "fargate" {
  cluster_name = aws_ecs_cluster.cluster.name

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
  }
}

# Gotta loop those IDs and filter for "Elastic network interface"
# network_interfaces = {
#   "filter" = toset(null) /* of object */
#   "id" = "us-east-1"
#   "ids" = tolist([
#     "eni-02f791f9734b39646",
#     "eni-0049d5ed19c37fa6e",
#     "eni-073e7b524fb3e9091",
#     "eni-097aade40750295e2",
#     "eni-0f7e5f7de7ff909ad",
#     "eni-0a3b5430e14eac346",
#     "eni-0a17ab38bad962386",
#     "eni-09fddacb2fa853f35",
#     "eni-04874b8bb409333b1",
#     "eni-0a1015cab993b7af9",
#     "eni-0b877a66ec1d25ead",
#     "eni-03ba959f37ae9dfa5",
#     "eni-0887239ccee0b7f21",
#     "eni-096cece683417b9bb",
#     "eni-02d50a03089daf965",
#   ])
#   "tags" = tomap(null) /* of string */
#   "timeouts" = null /* object */
# }

# data "aws_network_interfaces" "all" {}

# data "aws_network_interface" "interface_tags" {
#   depends_on = [ aws_ecs_service.service ]

#   filter {
#     name   = "tag:aws:ecs:serviceName"
#     values = ["disposable-chat-server"]
#   }
# }
