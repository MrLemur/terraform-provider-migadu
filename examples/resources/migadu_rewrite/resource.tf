resource "migadu_rewrite" "example" {
  domain_name     = "example.com"
  name            = "catch-support"
  local_part_rule = "support+"
  order_num       = 10
  destinations    = ["helpdesk@example.com"]
}

# Example with wildcard rule
resource "migadu_rewrite" "wildcard" {
  domain_name     = "example.com"
  name            = "team-wildcard"
  local_part_rule = "team-*"
  order_num       = 20
  destinations    = ["team@example.com", "manager@example.com"]
}
