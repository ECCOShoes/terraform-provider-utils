---
page_title: "readsops function - utils"
subcategory: ""
description: |-
  Decrypt SOPS-encrypted content.
---

# function: readsops

Decrypts [SOPS](https://github.com/getsops/sops)-encrypted content inline. The provider must have access to the appropriate decryption key (AWS KMS, GCP KMS, Azure Key Vault, age, or PGP).

## Signature

```text
readsops(encrypted string, format string) string
```

## Arguments

1. `encrypted` (String) — The SOPS-encrypted content (e.g. from `file()`).
2. `format` (String) — The format of the encrypted content. One of: `json`, `yaml`, `ini`, `dotenv`, `binary`.

## Example Usage

```terraform
# Decrypt a SOPS-encrypted JSON file
output "secrets" {
  value     = provider::utils::readsops(file("secrets.enc.json"), "json")
  sensitive = true
}

# Decrypt a SOPS-encrypted YAML file
output "config" {
  value     = provider::utils::readsops(file("config.enc.yaml"), "yaml")
  sensitive = true
}
```
