output "plan_role_arn" {
  value = aws_iam_role.terraform_plan.arn
}

output "apply_role_arn" {
  value = aws_iam_role.terraform_apply.arn
}
