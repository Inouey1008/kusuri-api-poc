module "terraform_actions_role" {
  source = "../../../modules/terraform_actions_role"

  shared_prefix                    = local.shared_prefix
  repository_name                  = local.repository_name
  apply_branches                   = local.apply_branches
  oidc_provider_arn                = module.github_actions_oidc_provider.oidc_provider_arn
  iam_policy_terraform_backend_arn = module.terraform_backend.iam_policy_terraform_backend_arn
}

output "plan_role_arn" {
  description = "GitHub Actions の secrets/vars に設定する値"
  value       = module.terraform_actions_role.plan_role_arn
}

output "apply_role_arn" {
  value = module.terraform_actions_role.apply_role_arn
}
