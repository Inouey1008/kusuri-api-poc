module "api" {
  source = "../../modules/api"

  shared_prefix = local.shared_prefix
  environment   = local.env
  memory_size   = local.memory_size

  slack_team_id    = var.slack_team_id
  slack_channel_id = var.slack_channel_id

  docs_user     = var.docs_user
  docs_password = var.docs_password
}

output "function_name" {
  description = "make deploy に渡す関数名"
  value       = module.api.function_name
}

output "alias_name" {
  value = module.api.alias_name
}

output "api_endpoint" {
  value = module.api.api_endpoint
}
