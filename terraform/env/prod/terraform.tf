// bucket は initial_setup/env/prod を apply して作る

terraform {
  required_version = "~> 1.10"

  backend "s3" {
    bucket       = "kusuri-api-poc-prod-terraform-backend"
    key          = "kusuri-api-poc/terraform.tfstate"
    region       = "ap-northeast-1"
    use_lockfile = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.0"
    }
  }
}
