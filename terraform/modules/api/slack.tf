# 認可はコンソールでの手動作業。ID は TF_VAR_ で注入する
locals {
  slack_enabled = var.slack_team_id != "" && var.slack_channel_id != ""
}

resource "aws_iam_role" "chatbot" {
  count = local.slack_enabled ? 1 : 0

  name = "${var.shared_prefix}-chatbot"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRole"
      Principal = { Service = "chatbot.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "chatbot_readonly" {
  count = local.slack_enabled ? 1 : 0

  role       = aws_iam_role.chatbot[0].name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess"
}

resource "aws_chatbot_slack_channel_configuration" "alerts" {
  count = local.slack_enabled ? 1 : 0

  configuration_name = "${var.shared_prefix}-alerts"
  iam_role_arn       = aws_iam_role.chatbot[0].arn
  slack_team_id      = var.slack_team_id
  slack_channel_id   = var.slack_channel_id
  sns_topic_arns     = [aws_sns_topic.alerts.arn]
}
