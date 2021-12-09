resource "aws_ecr_repository" "one-click-cosmos-testnet-repo" {
  name                 = "one-click-cosmos-testnet-repo"
  image_tag_mutability = "IMMUTABLE"
}

resource "aws_ecr_repository_policy" "one-click-cosmos-testnet-policy" {
  repository = aws_ecr_repository.one-click-cosmos-testnet-repo.name
  policy     = <<EOF
  {
    "Version": "2008-10-17",
    "Statement": [
      {
        "Sid": "adds full ecr access to the demo repository",
        "Effect": "Allow",
        "Principal": "*",
        "Action": [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:CompleteLayerUpload",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetLifecyclePolicy",
          "ecr:InitiateLayerUpload",
          "ecr:PutImage",
          "ecr:UploadLayerPart"
        ]
      }
    ]
  }
  EOF
}

resource "aws_ecs_cluster" "testnet-cluster" {
  name = "testnet-cluster"
}

resource "aws_vpc" "vpc" {
    cidr_block = "10.0.0.0/16"
    enable_dns_support   = true
    enable_dns_hostnames = true
    tags       = {
        Name = "Terraform VPC"
    }
}

resource "aws_security_group" "ecs_sg" {
    ingress {
        from_port       = 22
        to_port         = 22
        protocol        = "tcp"
        cidr_blocks     = ["0.0.0.0/0"]
    }
    ingress {
        from_port       = 26656
        to_port         = 26656
        protocol        = "tcp"
        cidr_blocks     = ["0.0.0.0/0"]
    }

    ingress {
        from_port       = 443
        to_port         = 443
        protocol        = "tcp"
        cidr_blocks     = ["0.0.0.0/0"]
    }

    egress {
        from_port       = 0
        to_port         = 65535
        protocol        = "tcp"
        cidr_blocks     = ["0.0.0.0/0"]
    }
}


resource "aws_launch_configuration" "ecs" {
  name                        = "testnet-cluster"
  image_id                    = "ami-00c7dbcc1310fd066"
  instance_type               = "t2.small"
  security_groups             = [aws_security_group.ecs_sg.id]
  associate_public_ip_address = true
  user_data                   = "#!/bin/bash\necho ECS_CLUSTER='testnet-cluster' > /etc/ecs/ecs.config"
}

resource "aws_autoscaling_group" "ecs-cluster" {
  availability_zones   = ["ap-south-1a", "ap-south-1b","ap-south-1c"]
  name                 = "testnet-cluster-auto-scaling-group"
  min_size             = 3
  max_size             = 3
  desired_capacity     = 3
  health_check_type    = "EC2"
  launch_configuration = aws_launch_configuration.ecs.name
}

resource "aws_ecs_service" "testnet-cluster-service" {
  name            = "testnet-app"
  cluster         = aws_ecs_cluster.testnet-cluster.id
  task_definition = aws_ecs_task_definition.testnet-ecs-task-definition.arn
  launch_type     = "EC2"
  desired_count = 3
}

resource "aws_ecs_task_definition" "testnet-ecs-task-definition" {
  family                   = "testnet-ecs-task-definition"
  network_mode             = "host"
  requires_compatibilities = ["EC2"]
  memory                   = "500"
  volume {
      name = "validator-config-volume"
      host_path = "/validator-config"
  }
  container_definitions = jsonencode([
    {
        name  = "testnet-container"
        image = "${aws_ecr_repository.one-click-cosmos-testnet-repo.repository_url}"
        memory = 128
        cpu = 128
        essential = true
        entryPoint = null
    }
  ])
}