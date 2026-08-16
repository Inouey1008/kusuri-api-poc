locals {
  service_name  = "kusuri-api-poc"
  env           = "dev"
  shared_prefix = "${local.service_name}-${local.env}"
  region        = "ap-northeast-1"

  # コールドスタート計測時はこの値を変えて apply し直す
  memory_size = 512
}
