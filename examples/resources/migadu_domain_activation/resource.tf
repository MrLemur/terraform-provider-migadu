resource "migadu_domain_activation" "example" {
  domain_name = "example.com"
}

# Typical usage: activate after creating the domain
resource "migadu_domain" "example" {
  name = "example.com"
}

resource "migadu_domain_activation" "example_activated" {
  domain_name = migadu_domain.example.name
}
