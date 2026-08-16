variable "shared_prefix" {
  type = string
}

variable "repository_name" {
  type        = string
  description = "owner/repo 形式"
}

variable "apply_branches" {
  type        = list(string)
  description = "apply を許可するブランチ (例: release/dev)"
}

variable "oidc_provider_arn" {
  type = string
}

variable "iam_policy_terraform_backend_arn" {
  type = string
}
