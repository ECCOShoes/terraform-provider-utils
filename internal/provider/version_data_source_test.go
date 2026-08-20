// Copyright (c) ECCO A/S
// SPDX-License-Identifier: MIT

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestVersionDataSource_Release(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "utils_version" "test" {
					base_version    = "1.2"
					rev_number      = 3
					build_number    = 42
					release_channel = "release"
				}
				output "dotnet_app_version" { value = data.utils_version.test.dotnet.app_version }
				output "dotnet_info_version" { value = data.utils_version.test.dotnet.info_version }
				output "debian_version" { value = data.utils_version.test.debian.version }
				output "nuget_version" { value = data.utils_version.test.nuget.version }
				output "npm_version" { value = data.utils_version.test.npm.version }
				output "docker_version" { value = data.utils_version.test.docker.version }
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("dotnet_app_version", knownvalue.StringExact("1.2.3.42")),
					statecheck.ExpectKnownOutputValue("dotnet_info_version", knownvalue.StringExact("1.2.3")),
					statecheck.ExpectKnownOutputValue("debian_version", knownvalue.StringExact("1.2.3-1")),
					statecheck.ExpectKnownOutputValue("nuget_version", knownvalue.StringExact("1.2.3")),
					statecheck.ExpectKnownOutputValue("npm_version", knownvalue.StringExact("1.2.3")),
					statecheck.ExpectKnownOutputValue("docker_version", knownvalue.StringExact("1.2.3")),
				},
			},
		},
	})
}

func TestVersionDataSource_ReleaseWithGitSha(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "utils_version" "test" {
					base_version    = "1.2"
					rev_number      = 3
					build_number    = 42
					git_sha         = "abcdef1"
					release_channel = "release"
				}
				output "dotnet_app_version" { value = data.utils_version.test.dotnet.app_version }
				output "dotnet_info_version" { value = data.utils_version.test.dotnet.info_version }
				output "debian_version" { value = data.utils_version.test.debian.version }
				output "nuget_version" { value = data.utils_version.test.nuget.version }
				output "npm_version" { value = data.utils_version.test.npm.version }
				output "docker_version" { value = data.utils_version.test.docker.version }
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					// git_sha is ignored by the purely-numeric dotnet app_version, and by docker on the release channel.
					statecheck.ExpectKnownOutputValue("dotnet_app_version", knownvalue.StringExact("1.2.3.42")),
					statecheck.ExpectKnownOutputValue("dotnet_info_version", knownvalue.StringExact("1.2.3-abcdef1")),
					statecheck.ExpectKnownOutputValue("debian_version", knownvalue.StringExact("1.2.3+gitabcdef1-1")),
					statecheck.ExpectKnownOutputValue("nuget_version", knownvalue.StringExact("1.2.3+abcdef1")),
					statecheck.ExpectKnownOutputValue("npm_version", knownvalue.StringExact("1.2.3+abcdef1")),
					statecheck.ExpectKnownOutputValue("docker_version", knownvalue.StringExact("1.2.3")),
				},
			},
		},
	})
}

func TestVersionDataSource_AlphaWithGitSha(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "utils_version" "test" {
					base_version    = "1.2"
					rev_number      = 3
					build_number    = 42
					git_sha         = "abcdef1"
					release_channel = "alpha"
					debian_date     = "20240101"
					arch            = "amd64"
				}
				output "dotnet_app_version" { value = data.utils_version.test.dotnet.app_version }
				output "dotnet_info_version" { value = data.utils_version.test.dotnet.info_version }
				output "debian_version" { value = data.utils_version.test.debian.version }
				output "debian_arch" { value = data.utils_version.test.debian.arch }
				output "nuget_version" { value = data.utils_version.test.nuget.version }
				output "npm_version" { value = data.utils_version.test.npm.version }
				output "docker_version" { value = data.utils_version.test.docker.version }
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("dotnet_app_version", knownvalue.StringExact("1.2.3.42")),
					statecheck.ExpectKnownOutputValue("dotnet_info_version", knownvalue.StringExact("1.2-alpha-abcdef1")),
					statecheck.ExpectKnownOutputValue("debian_version", knownvalue.StringExact("1.2.3~alpha~git20240101.abcdef1-1")),
					statecheck.ExpectKnownOutputValue("debian_arch", knownvalue.StringExact("amd64")),
					statecheck.ExpectKnownOutputValue("nuget_version", knownvalue.StringExact("1.2.3-alpha.42+abcdef1")),
					statecheck.ExpectKnownOutputValue("npm_version", knownvalue.StringExact("1.2.3-alpha.42+abcdef1")),
					statecheck.ExpectKnownOutputValue("docker_version", knownvalue.StringExact("1.2.3-alpha.42-abcdef1")),
				},
			},
		},
	})
}

func TestVersionDataSource_BetaWithoutGitSha(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "utils_version" "test" {
					base_version    = "1.2"
					rev_number      = 3
					build_number    = 42
					release_channel = "beta"
					debian_date     = "20240101"
					debian_revision = "2"
				}
				output "dotnet_info_version" { value = data.utils_version.test.dotnet.info_version }
				output "debian" { value = data.utils_version.test.debian }
				output "nuget_version" { value = data.utils_version.test.nuget.version }
				output "docker_version" { value = data.utils_version.test.docker.version }
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("dotnet_info_version", knownvalue.StringExact("1.2-beta")),
					statecheck.ExpectKnownOutputValue("debian", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"version": knownvalue.StringExact("1.2.3~beta~git20240101-2"),
						"arch":    knownvalue.Null(),
					})),
					statecheck.ExpectKnownOutputValue("nuget_version", knownvalue.StringExact("1.2.3-beta.42")),
					statecheck.ExpectKnownOutputValue("docker_version", knownvalue.StringExact("1.2.3-beta.42")),
				},
			},
		},
	})
}

func TestVersionDataSource_DefaultReleaseChannel(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "utils_version" "test" {
					base_version = "1.2"
					rev_number   = 3
					build_number = 42
				}
				output "release_channel" { value = data.utils_version.test.release_channel }
				output "nuget_version" { value = data.utils_version.test.nuget.version }
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("release_channel", knownvalue.StringExact("alpha")),
					statecheck.ExpectKnownOutputValue("nuget_version", knownvalue.StringExact("1.2.3-alpha.42")),
				},
			},
		},
	})
}

func TestVersionDataSource_InvalidReleaseChannel(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "utils_version" "test" {
					base_version    = "1.2"
					rev_number      = 3
					build_number    = 42
					release_channel = "nightly"
				}
				`,
				ExpectError: regexp.MustCompile(`Invalid release_channel`),
			},
		},
	})
}

func TestVersionDataSource_InvalidBaseVersion(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "utils_version" "test" {
					base_version = "1.2.3"
					rev_number   = 3
					build_number = 42
				}
				`,
				ExpectError: regexp.MustCompile(`Invalid base_version`),
			},
		},
	})
}
