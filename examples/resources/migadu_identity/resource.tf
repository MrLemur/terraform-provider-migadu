resource "migadu_identity" "example" {
  domain_name = "example.com"
  mailbox     = "user"
  local_part  = "alias"
  name        = "User Alias"
}

# Example with restricted permissions
resource "migadu_identity" "send_only" {
  domain_name = "example.com"
  mailbox     = "user"
  local_part  = "noreply"
  name        = "No Reply"

  may_send               = true
  may_receive            = false
  may_access_imap        = false
  may_access_pop3        = false
  may_access_managesieve = false
}
