terraform {
  required_version = "~> 1.7"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }

  backend "s3" {
    bucket                  = "lucasmauro.com-terraform-states"
    key                     = "disposable-chat-server"
    region                  = "us-east-1"
    shared_credentials_file = "~/.aws/credentials"
  }
}

provider "aws" {
  region = "us-east-1"
}
