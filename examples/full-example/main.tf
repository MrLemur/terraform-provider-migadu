terraform {
  required_providers {
    migadu = {
      source = "MrLemur/migadu"
    }
  }
}

provider "migadu" {
  # It's recommended to use environment variables:
  # export MIGADU_USERNAME="admin@example.com"
  # export MIGADU_API_KEY="your-api-key"

  username = var.migadu_username
  api_key  = var.migadu_api_key
}

variable "migadu_username" {
  description = "Migadu admin username"
  type        = string
}

variable "migadu_api_key" {
  description = "Migadu API key"
  type        = string
  sensitive   = true
}

variable "domain_name" {
  description = "Domain name to manage"
  type        = string
  default     = "example.com"
}

variable "admin_mailbox_password" {
  description = "Password for the admin mailbox"
  type        = string
  sensitive   = true
}

variable "support_mailbox_password" {
  description = "Password for the support mailbox"
  type        = string
  sensitive   = true
}

# Create a primary mailbox
resource "migadu_mailbox" "admin" {
  domain_name = var.domain_name
  local_part  = "admin"
  name        = "Administrator"

  password_method = "password"
  password        = var.admin_mailbox_password

  may_send               = true
  may_receive            = true
  may_access_imap        = true
  may_access_pop3        = false
  may_access_managesieve = true
}

# Create a support mailbox
resource "migadu_mailbox" "support" {
  domain_name = var.domain_name
  local_part  = "support"
  name        = "Support Team"

  password_method = "password"
  password        = var.support_mailbox_password

  spam_action         = "folder"
  spam_aggressiveness = "moderate"
}

# Create a sales mailbox with invitation
resource "migadu_mailbox" "sales" {
  domain_name = var.domain_name
  local_part  = "sales"
  name        = "Sales Department"

  password_method         = "invitation"
  password_recovery_email = "admin@${var.domain_name}"
}

# Create an alias for info@ pointing to support
resource "migadu_alias" "info" {
  domain_name  = var.domain_name
  local_part   = "info"
  destinations = [migadu_mailbox.support.address]
}

# Create an alias for contact@ pointing to multiple mailboxes
resource "migadu_alias" "contact" {
  domain_name = var.domain_name
  local_part  = "contact"
  destinations = [
    migadu_mailbox.support.address,
    migadu_mailbox.sales.address,
  ]
}

# Create a temporary alias for a campaign
resource "migadu_alias" "campaign" {
  domain_name  = var.domain_name
  local_part   = "spring2026"
  destinations = [migadu_mailbox.sales.address]

  expireable         = true
  expires_on         = "2026-06-30"
  remove_upon_expiry = true
}

# Outputs
output "admin_email" {
  value       = migadu_mailbox.admin.address
  description = "Admin mailbox email address"
}

output "support_email" {
  value       = migadu_mailbox.support.address
  description = "Support mailbox email address"
}

output "info_alias" {
  value       = migadu_alias.info.address
  description = "Info alias email address"
}
