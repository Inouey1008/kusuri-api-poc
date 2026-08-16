resource "aws_lambda_function" "api" {
  function_name = "${var.shared_prefix}-api"
  role          = aws_iam_role.lambda.arn

  # provided.al2023 は bootstrap という名前の実行ファイルを起動する
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]

  # 関数の作成には何らかのコードが必要なため、空の zip を置いておく
  # 実際のコードは deploy ワークフローが update-function-code で反映する
  filename = data.archive_file.placeholder.output_path

  # 更新のたびにバージョンを発行し、切り戻せるようにする
  publish = true

  memory_size = var.memory_size
  timeout     = var.timeout

  lifecycle {
    # コードの更新は CLI が担うため、Terraform は差分を見ない
    ignore_changes = [filename, source_code_hash]
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_cloudwatch_log_group.lambda,
  ]
}

data "archive_file" "placeholder" {
  type        = "zip"
  output_path = "${path.module}/placeholder.zip"

  source {
    content  = "placeholder"
    filename = "bootstrap"
  }
}

# 公開するバージョンを指すポインタ。疎通確認が通った後に切り替える
resource "aws_lambda_alias" "current" {
  name             = "current"
  function_name    = aws_lambda_function.api.function_name
  function_version = aws_lambda_function.api.version

  lifecycle {
    # 切り替えは CLI が担うため、Terraform は差分を見ない
    ignore_changes = [function_version]
  }
}

# PoC のため認証なしで公開する。CORS は未設定 (ブラウザからは直接呼べない)
resource "aws_lambda_function_url" "api" {
  function_name      = aws_lambda_function.api.function_name
  qualifier          = aws_lambda_alias.current.name
  authorization_type = "NONE"
}

# 明示的に作らないと Lambda が保持期間なしで自動生成する
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.shared_prefix}-api"
  retention_in_days = var.log_retention_in_days
}
