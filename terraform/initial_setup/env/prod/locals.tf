locals {
  service_name  = "kusuri-api-poc"
  env           = "prod"
  shared_prefix = "${local.service_name}-${local.env}"
  region        = "ap-northeast-1"

  subject_prefix = "repo:Inouey1008@108938387/kusuri-api-poc@1327389768"
  apply_branches = ["release/prod"]
}
