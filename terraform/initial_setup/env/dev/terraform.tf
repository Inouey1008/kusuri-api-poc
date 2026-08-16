// tfstate の保存先そのものを作るため、この構成だけはローカル state で管理する

terraform {
  required_version = "~> 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}
