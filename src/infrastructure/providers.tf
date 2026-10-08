terraform {
  required_version = ">= 1.16.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.68.0"
    }
  }
}

provider "aws" {
  region = "ap-southeast-2"

  # to apply common tags across resources I create
  default_tags {
    tags = {
      ManagedBy = "terraform"
    }
  }
}
