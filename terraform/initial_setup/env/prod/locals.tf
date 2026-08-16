locals {
  service_name  = "kusuri-api-poc"
  env           = "prod"
  shared_prefix = "${local.service_name}-${local.env}"
  region        = "ap-northeast-1"

  repository_name = "Inouey1008/kusuri-api-poc"
  apply_branches  = ["release/prod"]
}
