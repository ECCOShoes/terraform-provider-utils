terraform {
  required_providers {
    utils = {
      source = "ECCOShoes/utils"
    }
  }
}

provider "utils" {}

# Release build: plain numeric versions everywhere.
data "utils_version" "release" {
  base_version    = "1.2"
  rev_number      = 3
  build_number    = 42
  release_channel = "release"
}

output "release_dotnet_app_version" {
  value = data.utils_version.release.dotnet.app_version # => "1.2.3.42"
}

output "release_debian_version" {
  value = data.utils_version.release.debian.version # => "1.2.3-1"
}

# Alpha build from a feature branch: prerelease suffixes with commit sha.
data "utils_version" "alpha" {
  base_version    = "1.2"
  rev_number      = 3
  build_number    = 42
  git_sha         = "abcdef1"
  release_channel = "alpha"
  arch            = "amd64"
}

output "alpha_dotnet_info_version" {
  value = data.utils_version.alpha.dotnet.info_version # => "1.2-alpha-abcdef1"
}

output "alpha_nuget_version" {
  value = data.utils_version.alpha.nuget.version # => "1.2.3-alpha.42+abcdef1"
}

output "alpha_npm_version" {
  value = data.utils_version.alpha.npm.version # => "1.2.3-alpha.42+abcdef1"
}

output "alpha_docker_version" {
  value = data.utils_version.alpha.docker.version # => "1.2.3-alpha.42-abcdef1"
}

output "alpha_debian_version" {
  value = data.utils_version.alpha.debian.version # => "1.2.3~alpha~git<today>.abcdef1-1"
}
