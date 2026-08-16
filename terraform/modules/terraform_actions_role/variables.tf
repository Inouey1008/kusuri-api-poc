variable "shared_prefix" {
  type = string
}

# GitHub は名前変更に耐えるよう sub に数値 ID を含める。値は以下で取得する
#   gh api /repos/<owner>/<repo>/actions/oidc/customization/sub
variable "subject_prefix" {
  type        = string
  description = "OIDC トークンの sub のプレフィックス (例: repo:owner@123/repo@456)"
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
