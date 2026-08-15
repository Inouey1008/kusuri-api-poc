module "api" {
  source = "../../modules/api"

  shared_prefix = local.shared_prefix
  memory_size   = local.memory_size
}

output "function_name" {
  description = "make deploy に渡す関数名"
  value       = module.api.function_name
}

output "function_url" {
  value = module.api.function_url
}
