---
page_title: "deepmerge function - utils"
subcategory: ""
description: |-
  Deep-merge multiple objects into a single value.
---

# function: deepmerge

Recursively merges multiple objects into a single dynamic value.

**Merge behavior:**

- **Objects** — merged recursively (nested keys are combined).
- **Arrays** — concatenated (elements from all arrays are joined).
- **Scalars** (strings, numbers, booleans) — the second value overwrites the first when `overwrite = true`.

When `overwrite = false`, conflicting keys with primitive values cause an error.

A single list argument is also accepted — all items in that list will be merged together. This is useful when the set of objects to merge is computed dynamically.

## Signature

```text
deepmerge(overwrite bool, objects dynamic...) dynamic
```

## Arguments

1. `overwrite` (Boolean) — Allow overwriting conflicting primitive keys. Set to `false` to error on conflicts, `true` to let the last value win.
2. `objects` (Dynamic, variadic) — Two or more objects to merge, **or** a single list containing objects to merge.

## Example Usage

```terraform
# Merge two objects (no overwrite)
output "merged" {
  value = provider::utils::deepmerge(false,
    { hello = { world = "test" } },
    { hello = { qwe = "asd" } }
  )
}
# => { hello = { world = "test", qwe = "asd" } }

# Overwrite scalar conflicts
output "overwritten" {
  value = provider::utils::deepmerge(true,
    { key = "value1" },
    { key = "value2" }
  )
}
# => { key = "value2" }

# Merge a dynamic list of objects
locals {
  layers = [
    { api = { endpoint = "https://api.example.com", version = "v1" } },
    { api = { timeout = 30 } },
    { features = ["logging", "monitoring"] }
  ]
}

output "from_list" {
  value = provider::utils::deepmerge(true, local.layers)
}
```
