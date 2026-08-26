locals {
  # ロールを引き受けられる GitHub Actions の実行文脈
  plan_subjects  = ["${var.subject_prefix}:pull_request"]
  apply_subjects = [for branch in var.apply_branches : "${var.subject_prefix}:ref:refs/heads/${branch}"]
}

resource "aws_iam_role" "terraform_plan" {
  name = "${var.shared_prefix}-terraform-plan"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRoleWithWebIdentity"
      Principal = { Federated = var.oidc_provider_arn }
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          "token.actions.githubusercontent.com:sub" = local.plan_subjects
        }
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "terraform_plan_readonly" {
  role       = aws_iam_role.terraform_plan.name
  policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"
}

# plan もロック取得と state 更新のため書き込みが要る
resource "aws_iam_role_policy_attachment" "terraform_plan_backend" {
  role       = aws_iam_role.terraform_plan.name
  policy_arn = var.iam_policy_terraform_backend_arn
}

resource "aws_iam_role" "terraform_apply" {
  name = "${var.shared_prefix}-terraform-apply"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRoleWithWebIdentity"
      Principal = { Federated = var.oidc_provider_arn }
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          "token.actions.githubusercontent.com:sub" = local.apply_subjects
        }
      }
    }]
  })
}

# 作成対象のサービスに限定する
data "aws_iam_policy_document" "terraform_apply" {
  statement {
    actions   = ["lambda:*"]
    resources = ["*"]
  }

  statement {
    actions   = ["logs:*"]
    resources = ["*"]
  }

  # Lambda の暖機スケジュール
  statement {
    actions   = ["scheduler:*"]
    resources = ["*"]
  }

  statement {
    actions   = ["apigateway:*"]
    resources = ["*"]
  }

  statement {
    actions = [
      "cloudwatch:*",
      "sns:*",
      "chatbot:*",
    ]
    resources = ["*"]
  }

  # Lambda 実行ロールと Scheduler 実行ロールの管理
  statement {
    actions = [
      "iam:CreateRole",
      "iam:DeleteRole",
      "iam:GetRole",
      "iam:ListRolePolicies",
      "iam:ListAttachedRolePolicies",
      "iam:ListInstanceProfilesForRole",
      "iam:AttachRolePolicy",
      "iam:DetachRolePolicy",
      "iam:PutRolePolicy",
      "iam:GetRolePolicy",
      "iam:DeleteRolePolicy",
      "iam:TagRole",
      "iam:UntagRole",
      "iam:PassRole",
      "iam:UpdateAssumeRolePolicy",
      "iam:UpdateRole",
      "iam:UpdateRoleDescription",
    ]
    resources = ["*"]
  }

  # Chatbot は初回に、サービスにリンクされたロールを自分で作る
  statement {
    actions   = ["iam:CreateServiceLinkedRole"]
    resources = ["*"]

    condition {
      test     = "StringEquals"
      variable = "iam:AWSServiceName"
      values   = ["management.chatbot.amazonaws.com"]
    }
  }

  # 仕様書の資格情報を SSM で管理する
  statement {
    actions = [
      "ssm:PutParameter",
      "ssm:DeleteParameter",
      "ssm:AddTagsToResource",
      "ssm:RemoveTagsFromResource",
    ]
    resources = ["*"]
  }

  statement {
    actions   = ["kms:Encrypt", "kms:GenerateDataKey"]
    resources = ["*"]

    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["ssm.${var.region}.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "terraform_apply" {
  name   = "${var.shared_prefix}-terraform-apply"
  policy = data.aws_iam_policy_document.terraform_apply.json
}

resource "aws_iam_role_policy_attachment" "terraform_apply" {
  role       = aws_iam_role.terraform_apply.name
  policy_arn = aws_iam_policy.terraform_apply.arn
}

# plan の差分計算に既存リソースの参照が要る
resource "aws_iam_role_policy_attachment" "terraform_apply_readonly" {
  role       = aws_iam_role.terraform_apply.name
  policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"
}

resource "aws_iam_role_policy_attachment" "terraform_apply_backend" {
  role       = aws_iam_role.terraform_apply.name
  policy_arn = var.iam_policy_terraform_backend_arn
}

# ReadOnlyAccess は SSM の取得までで、SecureString の復号は許可しない
data "aws_iam_policy_document" "ssm_decrypt" {
  statement {
    actions   = ["kms:Decrypt"]
    resources = ["*"]

    # SSM 経由の復号だけに絞る
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["ssm.${var.region}.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "ssm_decrypt" {
  name   = "${var.shared_prefix}-ssm-decrypt"
  policy = data.aws_iam_policy_document.ssm_decrypt.json
}

resource "aws_iam_role_policy_attachment" "terraform_plan_ssm_decrypt" {
  role       = aws_iam_role.terraform_plan.name
  policy_arn = aws_iam_policy.ssm_decrypt.arn
}

resource "aws_iam_role_policy_attachment" "terraform_apply_ssm_decrypt" {
  role       = aws_iam_role.terraform_apply.name
  policy_arn = aws_iam_policy.ssm_decrypt.arn
}
