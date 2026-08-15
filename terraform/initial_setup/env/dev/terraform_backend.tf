module "terraform_backend" {
  source        = "../../../modules/terraform_backend"
  shared_prefix = local.shared_prefix
}

output "bucket_name" {
  description = "env/dev の terraform.tf に設定する値"
  value       = module.terraform_backend.bucket_name
}
