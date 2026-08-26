module "api" {
  source = "../../modules/api"

  shared_prefix = local.shared_prefix
  environment   = local.env
  memory_size   = local.memory_size

  slack_team_id    = local.slack_enabled ? data.aws_ssm_parameter.slack_team_id[0].value : ""
  slack_channel_id = local.slack_enabled ? data.aws_ssm_parameter.slack_channel_id[0].value : ""

  docs_user     = aws_ssm_parameter.docs_user.value
  docs_password = aws_ssm_parameter.docs_password.value
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
