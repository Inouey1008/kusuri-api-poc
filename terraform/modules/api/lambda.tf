resource "aws_lambda_function" "api" {
  function_name = "${var.shared_prefix}-api"
  role          = aws_iam_role.lambda.arn

  # provided.al2023 は bootstrap という名前の実行ファイルを起動する
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]

  # 作成にはコードが必要。実体は deploy が入れる
  filename = data.archive_file.placeholder.output_path

  publish = true

  memory_size = var.memory_size
  timeout     = var.timeout

  # 流量は API Gateway が絞る。ここは保険
  reserved_concurrent_executions = var.reserved_concurrency

  environment {
    variables = {
      ENVIRONMENT = var.environment

      DOCS_USER     = var.docs_user
      DOCS_PASSWORD = var.docs_password

      # アプリが必須にしているが、Lambda では使われない
      PORT = "8080"
    }
  }

  lifecycle {
    # 更新は CLI が担う
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

# 疎通確認が通った後に切り替える
resource "aws_lambda_alias" "current" {
  name             = "current"
  function_name    = aws_lambda_function.api.function_name
  function_version = aws_lambda_function.api.version

  lifecycle {
    # 切り替えは CLI が担う
    ignore_changes = [function_version]
  }
}

# 作らないと保持期間なしで自動生成される
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.shared_prefix}-api"
  retention_in_days = var.log_retention_in_days
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "${var.shared_prefix}-api-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

# 外部リソースを参照しないため Logs 書き込みのみ
resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
