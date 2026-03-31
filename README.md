# Terraform Provider Utils

A Terraform provider that exposes utility functions for data manipulation. Built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

## Functions

| Function | Description |
|----------|-------------|
| `deepmerge` | Deep-merge multiple objects/maps into one. Recursively merges nested objects, concatenates arrays, and optionally overwrites conflicting scalar keys. |
| `readsops` | Decrypt [SOPS](https://github.com/getsops/sops)-encrypted content inline. Supports `json`, `yaml`, `ini`, `dotenv`, and `binary` formats. |
| `jsonexpand` | Expand a flat map into a nested object structure using `__` as the key separator. Optionally converts numeric-keyed objects into arrays. |
| `jsonflatten` | Flatten a nested object/map (including arrays) into a single-level `map(string)` with keys joined by `__`. |

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.8 (provider-defined functions require 1.8+)
- [Go](https://golang.org/doc/install) >= 1.24 (only for building from source)

## Usage

```hcl
terraform {
  required_providers {
    utils = {
      source = "ECCOShoes/utils"
    }
  }
}

provider "utils" {}
```

### `deepmerge`

Recursively merge two or more objects. Objects are merged key-by-key, arrays are concatenated, and scalar conflicts are controlled by the `overwrite` flag.

```hcl
# Merge multiple objects
output "merged" {
  value = provider::utils::deepmerge(true,
    { api = { endpoint = "https://api.example.com", version = "v1" } },
    { api = { timeout = 30 } },
    { features = ["logging"] }
  )
}

# Merge a list of objects (useful with for expressions)
locals {
  configs = [
    { database = { host = "localhost", port = 5432 } },
    { database = { name = "myapp" } },
    { cache = { enabled = true } }
  ]
}

output "merged_list" {
  value = provider::utils::deepmerge(true, local.configs)
}
```

### `readsops`

Decrypt SOPS-encrypted content. The provider must have access to the appropriate key (AWS KMS, GCP KMS, Azure Key Vault, age, or PGP).

```hcl
output "secrets" {
  value     = provider::utils::readsops(file("secrets.enc.json"), "json")
  sensitive = true
}
```

### `jsonexpand`

Expand a flat map with `__`-separated keys into a nested structure. Pass `true` as the optional second argument to convert objects with numeric keys (`0`, `1`, `2`, …) into arrays.

```hcl
locals {
  flat = {
    "api__endpoint" = "https://api.example.com"
    "api__timeout"  = "30"
    "features__0"   = "logging"
    "features__1"   = "monitoring"
  }
}

output "nested" {
  value = provider::utils::jsonexpand(local.flat)
}

output "nested_with_arrays" {
  value = provider::utils::jsonexpand(local.flat, true)
}
```

### `jsonflatten`

Flatten a nested object into a single-level `map(string)`. Arrays are flattened with numeric indices.

```hcl
locals {
  nested = {
    api = {
      endpoint = "https://api.example.com"
      timeout  = 30
    }
    features = ["logging", "monitoring"]
  }
}

output "flat" {
  value = provider::utils::jsonflatten(local.nested)
  # => { "api__endpoint" = "https://api.example.com", "api__timeout" = "30", "features__0" = "logging", "features__1" = "monitoring" }
}
```

### Round-trip: flatten → expand

`jsonflatten` and `jsonexpand` are inverse operations, making it easy to transform, modify, and reconstruct nested structures:

```hcl
locals {
  original = {
    hello = { world = "test", foo = "bar" }
    arr   = ["x", "y"]
    top   = "level"
  }
}

output "roundtrip" {
  value = provider::utils::jsonexpand(
    provider::utils::jsonflatten(local.original),
    true
  )
}
```

## Building The Provider

```shell
go install
```

## Developing

```shell
# Run unit tests
make test

# Run acceptance tests (requires Terraform CLI)
make testacc

# Lint
make lint

# Generate documentation
make generate
```
