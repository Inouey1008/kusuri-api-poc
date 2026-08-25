locals {
  service_name  = "kusuri-api-poc"
  env           = "prod"
  shared_prefix = "${local.service_name}-${local.env}"
  region        = "ap-northeast-1"

  # false にすると SSM を読まず、Slack 連携も作らない
  slack_enabled = true

  # コールドスタート計測時はこの値を変えて apply し直す
  memory_size = 512
}
