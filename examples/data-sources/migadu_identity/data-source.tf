data "migadu_identity" "example" {
  domain_name = "example.com"
  mailbox     = "user"
  local_part  = "alias"
}
