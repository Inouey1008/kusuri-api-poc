# 仕様書の資格情報は Terraform が生成し、SSM に置く。
# 値を CI やコードに持たせずに済む。差し替えたい場合は SSM を直接書き換える
resource "random_password" "docs" {
  length = 32
}

resource "aws_ssm_parameter" "docs_user" {
  name  = "/${local.service_name}/${local.env}/docs_user"
  type  = "SecureString"
  value = "docs"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "docs_password" {
  name  = "/${local.service_name}/${local.env}/docs_password"
  type  = "SecureString"
  value = random_password.docs.result

  lifecycle {
    ignore_changes = [value]
  }
}

# Slack の識別子は実在する値のため生成できない。手作業で登録する (README を参照)
data "aws_ssm_parameter" "slack_team_id" {
  count = local.slack_enabled ? 1 : 0

  name = "/${local.service_name}/${local.env}/slack_team_id"
}

data "aws_ssm_parameter" "slack_channel_id" {
  count = local.slack_enabled ? 1 : 0

  name = "/${local.service_name}/${local.env}/slack_channel_id"
}
