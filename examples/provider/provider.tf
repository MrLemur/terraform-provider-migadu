terraform {
  required_providers {
    migadu = {
      source = "MrLemur/migadu"
    }
  }
}

provider "migadu" {
  username = "admin@example.com"
  api_key  = "your-api-key-here"
}
