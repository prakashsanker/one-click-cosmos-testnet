/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var dataTemplate = `
data "aws_ami" "ecs" {
	most_recent = true # get the latest version
  
	filter {
	  name = "name"
	  values = [
		"amzn2-ami-ecs-*"] # ECS optimized image
	}
  
	filter {
	  name = "virtualization-type"
	  values = [
		"hvm"]
	}
  
	owners = [
	  "amazon" # Only official images
	]
}
`

var template = `
provider "aws" {
	region = "ap-south-1"
}
resource "tls_private_key" "validator_private_key" {
	algorithm = "RSA"
	rsa_bits  = 4096
}

resource "aws_key_pair" "generated_key" {
	key_name   = "validator_key"
	public_key = tls_private_key.validator_private_key.public_key_openssh
	# provisioner "local-exec" {
	#   command = "echo '${tls_private_key.validator_private_key.private_key_pem}' > ./validator_key.pem"
	# }
}

resource "local_file" "ssh_key" {
	filename = "validator_key.pem"
	content = tls_private_key.validator_private_key.private_key_pem
}

data "aws_iam_policy_document" "ecs_instance_assume_role_policy" {
	statement {
	actions = ["sts:AssumeRole"]

	principals {
		type        = "Service"
		identifiers = ["ec2.amazonaws.com"]
	}
	}
}

resource "aws_iam_role" "ecs_instance_role" {
	name               = "ecs-instance-role-testnet"
	assume_role_policy = data.aws_iam_policy_document.ecs_instance_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role_policy" {
	role       = aws_iam_role.ecs_instance_role.name
	policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_profile" {
	name = "ecsInstanceRole-testnet"
	path = "/"
	role = aws_iam_role.ecs_instance_role.name
}

resource "aws_ecr_repository" "one-click-cosmos-testnet-repo" {
	name                 = "one-click-cosmos-testnet-repo"
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
	iam_instance_profile = aws_iam_instance_profile.ecs_instance_profile.name
	image_id                    = data.aws_ami.ecs.id
	key_name                    = "validator_key"
	instance_type               = "t2.micro"
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
	deployment_minimum_healthy_percent = 0
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
		host_path = "/home/ec2-user/validator-config"
	}
	container_definitions = jsonencode([
	{
		name  = "testnet-container"
		image = "${aws_ecr_repository.one-click-cosmos-testnet-repo.repository_url}"
		memory = 128
		cpu = 128
		essential = true
		entryPoint = null
		mountPoints = [{
			sourceVolume = "validator-config-volume"
			containerPath = "/validator-config"
			readOnly = false
		}]
	}
	])
}

`

func generateTerraformTemplate(fileName string, template string) {
	terraformTemplate, err := os.Create(fileName)

	if err != nil {
		fmt.Println(err)
	}

	defer terraformTemplate.Close()
	_, err = terraformTemplate.WriteString(template)
	if err != nil {
		fmt.Println(err)
	}
	terraformTemplate.Sync()
}

var generateTestInfra = &cobra.Command{
	Use:   "generate-test-infra",
	Short: "Create cloud validator set",
	Long:  `Run this command to generate the infrastructure for your validators in AWS`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generate Test Infra called")
		// check to see if the clis exist

		generateTerraformTemplate("./testnet-template.tf", template)
		generateTerraformTemplate("./data.tf", dataTemplate)
		terraformExecutable, terraformErr := exec.LookPath("terraform")
		if terraformErr != nil {
			fmt.Println("Please install terraform")
			return
		}

		_, awsErr := exec.LookPath("aws")
		if awsErr != nil {
			fmt.Println("Please install aws-cli and run one-click-cosmos-testnet configure")
			return
		}

		terraformInitCmd := &exec.Cmd{
			Path:   terraformExecutable,
			Args:   []string{terraformExecutable, "init"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := terraformInitCmd.Run(); err != nil {
			fmt.Println("error :", err)
		}

		terraformApplyCmd := &exec.Cmd{
			Path:   terraformExecutable,
			Args:   []string{terraformExecutable, "apply", "-auto-approve"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := terraformApplyCmd.Run(); err != nil {
			fmt.Println("error: ", err)
		}

		chmodExecutable, _ := exec.LookPath("chmod")
		chmodValidatorKeyPem := &exec.Cmd{
			Path:   chmodExecutable,
			Args:   []string{chmodExecutable, "400", "validator_key.pem"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err := chmodValidatorKeyPem.Run(); err != nil {
			fmt.Println("chmod validator error: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateTestInfra)
}
