terraform {
  required_providers {
    utils = {
      source = "ECCOShoes/utils"
    }
  }
}

provider "utils" {}

# Merge two objects without allowing overwrites
output "basic_merge" {
  value = provider::utils::deepmerge(false,
    { hello = { world = "test" } },
    { hello = { qwe = "asd" } }
  )
}

# Merge with overwrite enabled
output "overwrite_merge" {
  value = provider::utils::deepmerge(true,
    { key = "original" },
    { key = "updated" }
  )
}

# Merge a dynamic list of objects
locals {
  config_layers = [
    { api = { endpoint = "https://api.example.com", version = "v1" } },
    { api = { timeout = 30, retries = 3 } },
    { features = ["logging", "monitoring"] },
    { features = ["alerting"] }
  ]
}

output "merged_list" {
  value = provider::utils::deepmerge(true, local.config_layers)
}
