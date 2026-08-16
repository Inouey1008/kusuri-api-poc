provider "aws" {
  region = local.region

  default_tags {
    tags = {
      service = local.service_name
      env     = local.env
    }
  }
}
