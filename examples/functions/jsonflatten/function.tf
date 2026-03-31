terraform {
  required_providers {
    utils = {
      source = "ECCOShoes/utils"
    }
  }
}

provider "utils" {}

locals {
  nested = {
    level1 = {
      level2 = {
        level3  = "deep"
        another = "value"
      }
      sibling = true
    }
    top = "simple"
  }
}

output "flat" {
  value = provider::utils::jsonflatten(local.nested)
}
