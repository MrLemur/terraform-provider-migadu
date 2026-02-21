resource "migadu_alias" "example" {
  domain_name  = "example.com"
  local_part   = "info"
  destinations = ["hello@example.com", "support@example.com"]
}

# Example with expiration
resource "migadu_alias" "temporary" {
  domain_name  = "example.com"
  local_part   = "temp"
  destinations = ["admin@example.com"]

  expireable         = true
  expires_on         = "2026-12-31"
  remove_upon_expiry = true
}

