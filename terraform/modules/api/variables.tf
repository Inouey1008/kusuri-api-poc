variable "shared_prefix" {
  type = string
}

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

variable "warmup_enabled" {
  type    = bool
  default = true
}

variable "warmup_interval_minutes" {
  type    = number
  default = 5
}

variable "throttling_rate_limit" {
  type    = number
  default = 30
}

variable "throttling_burst_limit" {
  type    = number
  default = 60
}

variable "cors_allow_origins" {
  type    = list(string)
  default = []
}

# コールドスタート時でも throttling_rate_limit を捌けるだけ確保する
variable "reserved_concurrency" {
  type    = number
  default = 30
}

variable "alert_emails" {
  type    = list(string)
  default = []
}

variable "error_alarm_threshold" {
  type    = number
  default = 5
}

variable "request_spike_threshold" {
  type    = number
  default = 10000
}

variable "slack_team_id" {
  type    = string
  default = ""
}

variable "slack_channel_id" {
  type    = string
  default = ""
}

variable "docs_user" {
  type = string
}

variable "docs_password" {
  type      = string
  sensitive = true
}

variable "environment" {
  type = string
}
