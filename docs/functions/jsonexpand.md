---
page_title: "jsonexpand function - utils"
subcategory: ""
description: |-
  Expand a flat map into a nested object structure.
---

# function: jsonexpand

Expands a flat map into a nested object structure using `__` as the key separator.

When the optional `expand_arrays` flag is `true`, objects whose keys are all consecutive numeric indices (`0`, `1`, `2`, …) are automatically converted into arrays. Missing indices become `null`.

## Signature

```text
jsonexpand(obj dynamic, expand_arrays bool...) dynamic
```

## Arguments

1. `obj` (Dynamic) — A flat map/object to expand into a nested structure. All values must be strings.
2. `expand_arrays` (Boolean, optional variadic) — If `true`, numeric-keyed objects are expanded into arrays. Defaults to `false`.

## Example Usage

```terraform
locals {
  flat = {
    "api__endpoint" = "https://api.example.com"
    "api__timeout"  = "30"
    "features__0"   = "logging"
    "features__1"   = "monitoring"
  }
}

# Without array expansion — numeric keys stay as object keys
output "nested" {
  value = provider::utils::jsonexpand(local.flat)
}
# => { api = { endpoint = "https://api.example.com", timeout = "30" }, features = { "0" = "logging", "1" = "monitoring" } }

# With array expansion — numeric keys become array indices
output "nested_with_arrays" {
  value = provider::utils::jsonexpand(local.flat, true)
}
# => { api = { endpoint = "https://api.example.com", timeout = "30" }, features = ["logging", "monitoring"] }
```
