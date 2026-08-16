variable "shared_prefix" {
  type = string
}

# コールドスタート時間は割り当てられる CPU に左右され、CPU はメモリに比例する
variable "memory_size" {
  type    = number
  default = 512
}

variable "timeout" {
  type    = number
  default = 10
}

variable "log_retention_in_days" {
  type    = number
  default = 7
}
