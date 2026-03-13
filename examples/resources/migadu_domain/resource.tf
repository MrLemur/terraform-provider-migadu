resource "migadu_domain" "example" {
  name = "example.com"
}

# Example with custom settings
resource "migadu_domain" "custom" {
  name        = "custom.example.com"
  description = "Custom domain with spam filtering"

  spam_aggressiveness = "aggressive"
  greylisting_enabled = true

  catchall_destinations = ["admin@custom.example.com"]
}
