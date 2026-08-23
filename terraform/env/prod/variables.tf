variable "slack_team_id" {
  type    = string
  default = ""
}

variable "slack_channel_id" {
  type    = string
  default = ""
}

variable "docs_user" {
  type = string

  validation {
    condition     = var.docs_user != ""
    error_message = "TF_VAR_docs_user を設定してください (GitHub Secrets の DOCS_USER)。"
  }
}

variable "docs_password" {
  type      = string
  sensitive = true

  validation {
    condition     = var.docs_password != ""
    error_message = "TF_VAR_docs_password を設定してください (GitHub Secrets の DOCS_PASSWORD)。"
  }
}
