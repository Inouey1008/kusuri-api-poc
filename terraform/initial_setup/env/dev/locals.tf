locals {
  service_name  = "kusuri-api-poc"
  env           = "dev"
  shared_prefix = "${local.service_name}-${local.env}"
  region        = "ap-northeast-1"
}
