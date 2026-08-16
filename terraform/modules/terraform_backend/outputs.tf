output "bucket_name" {
  value = aws_s3_bucket.terraform_backend.bucket
}

output "iam_policy_terraform_backend_arn" {
  value = aws_iam_policy.terraform_backend.arn
}
