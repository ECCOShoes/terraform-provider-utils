---
page_title: "jsonflatten function - utils"
subcategory: ""
description: |-
  Flatten a nested object into a single-level map.
---

# function: jsonflatten

Flattens a nested object/map (and any nested arrays) into a single-level `map(string)` with keys separated by `__`. Arrays are flattened using numeric indices (`0`, `1`, `2`, …).

This is the inverse of [`jsonexpand`](jsonexpand.md) — you can round-trip data through `jsonflatten` → `jsonexpand` to reconstruct the original structure.

## Signature

```text
jsonflatten(obj dynamic) map(string)
```

## Arguments

1. `obj` (Dynamic) — A nested object/map or array to flatten.

## Example Usage

```terraform
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
# => {
#   "level1__level2__level3"  = "deep"
#   "level1__level2__another" = "value"
#   "level1__sibling"         = "true"
#   "top"                     = "simple"
# }

# Flatten arrays
output "flat_array" {
  value = provider::utils::jsonflatten({ items = ["a", "b", "c"] })
}
# => { "items__0" = "a", "items__1" = "b", "items__2" = "c" }

# Round-trip: flatten then expand
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
