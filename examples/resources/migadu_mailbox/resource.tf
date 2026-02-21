variable "mailbox_password" {
  description = "Password for the mailbox"
  type        = string
  sensitive   = true
}

resource "migadu_mailbox" "example" {
  domain_name = "example.com"
  local_part  = "hello"
  name        = "Hello User"

  # Using password authentication
  password_method = "password"
  password        = var.mailbox_password

  # Access permissions
  may_send               = true
  may_receive            = true
  may_access_imap        = true
  may_access_pop3        = false
  may_access_managesieve = true

  # Spam settings
  spam_action         = "folder"
  spam_aggressiveness = "moderate"
}

# Example with invitation method
resource "migadu_mailbox" "invitation" {
  domain_name = "example.com"
  local_part  = "newuser"
  name        = "New User"

  password_method         = "invitation"
  password_recovery_email = "recovery@otherdomain.com"
}
