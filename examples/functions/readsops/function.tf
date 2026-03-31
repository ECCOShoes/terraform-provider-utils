terraform {
  required_providers {
    utils = {
      source = "ECCOShoes/utils"
    }
  }
}

provider "utils" {}

# Decrypt a SOPS-encrypted JSON file
output "secrets" {
  value     = provider::utils::readsops(file("secrets.enc.json"), "json")
  sensitive = true
}
