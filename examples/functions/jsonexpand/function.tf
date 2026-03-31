terraform {
  required_providers {
    utils = {
      source = "ECCOShoes/utils"
    }
  }
}

provider "utils" {}

locals {
  flat = {
    "api__endpoint" = "https://api.example.com"
    "api__timeout"  = "30"
    "features__0"   = "logging"
    "features__1"   = "monitoring"
  }
}

# Expand without array conversion
output "nested" {
  value = provider::utils::jsonexpand(local.flat)
}

# Expand with array conversion enabled
output "nested_with_arrays" {
  value = provider::utils::jsonexpand(local.flat, true)
}
