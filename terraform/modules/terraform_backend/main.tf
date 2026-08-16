resource "aws_s3_bucket" "terraform_backend" {
  bucket = "${var.shared_prefix}-terraform-backend"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_backend" {
  bucket = aws_s3_bucket.terraform_backend.bucket
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_versioning" "terraform_backend" {
  bucket = aws_s3_bucket.terraform_backend.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_backend" {
  bucket                  = aws_s3_bucket.terraform_backend.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "terraform_backend" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.terraform_backend.arn]
  }

  statement {
    actions   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"]
    resources = ["${aws_s3_bucket.terraform_backend.arn}/*"]
  }
}

resource "aws_iam_policy" "terraform_backend" {
  name   = "${var.shared_prefix}-terraform-backend-policy"
  policy = data.aws_iam_policy_document.terraform_backend.json
}
