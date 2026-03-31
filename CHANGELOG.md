# Changelog

## 1.0.0 (Unreleased)

FEATURES:

* **New Function:** `deepmerge` — recursively merge multiple objects/maps with configurable overwrite behavior. Supports variadic objects and single-list input.
* **New Function:** `readsops` — decrypt SOPS-encrypted content inline. Supports `json`, `yaml`, `ini`, `dotenv`, and `binary` formats.
* **New Function:** `jsonexpand` — expand a flat `__`-separated map into a nested object structure. Optional array expansion for numeric keys.
* **New Function:** `jsonflatten` — flatten a nested object/map (including arrays) into a single-level `map(string)` with `__`-separated keys.

IMPROVEMENTS:

* Added GitHub Actions workflows for CI (tests, lint) and release (GoReleaser multiplatform builds).
* Added comprehensive unit and acceptance tests for all functions.
* Updated documentation and examples.
