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

func TestJsonflattenFunction_Known(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
				    tst = {
						hello = {
							world = "testvalue"
						}
					}
				}
				output "test" {
					value = provider::utils::jsonflatten(local.tst)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"hello__world": knownvalue.StringExact("testvalue"),
						}),
					),
				},
			},
		},
	})
}

func TestJsonflattenFunction_Null(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::utils::jsonflatten(null)
				}
				`,
				// The parameter does not enable AllowNullValue
				ExpectError: regexp.MustCompile(`argument must not be null`),
			},
		},
	})
}

func TestJsonflattenFunction_Array(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::utils::jsonflatten(["test", "value"])
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"0": knownvalue.StringExact("test"),
							"1": knownvalue.StringExact("value"),
						}),
					),
				},
			},
		},
	})
}

func TestJsonflattenFunction_NestedArray(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					tst = { hello = ["d", "e"] }
				}
				output "test" {
					value = provider::utils::jsonflatten(local.tst)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"hello__0": knownvalue.StringExact("d"),
							"hello__1": knownvalue.StringExact("e"),
						}),
					),
				},
			},
		},
	})
}

func TestJsonflattenFunction_Scalar(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::utils::jsonflatten("testvalue")
				}
				`,
				ExpectError: regexp.MustCompile(`Input must be a map/object or array`),
			},
		},
	})
}

func TestJsonflattenFunction_MultipleKeys(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
				    tst = {
						hello = {
							world = "test"
							asd = 45
						}
					}
				}
				output "test" {
					value = provider::utils::jsonflatten(local.tst)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"hello__world": knownvalue.StringExact("test"),
							"hello__asd":   knownvalue.StringExact("45"),
						}),
					),
				},
			},
		},
	})
}

func TestJsonflattenFunction_Complex(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
				    tst = {
						level1 = {
							level2 = {
								level3 = "deep"
								another = "value"
							}
							sibling = true
						}
						top = "simple"
					}
				}
				output "test" {
					value = provider::utils::jsonflatten(local.tst)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.MapExact(map[string]knownvalue.Check{
							"level1__level2__level3":  knownvalue.StringExact("deep"),
							"level1__level2__another": knownvalue.StringExact("value"),
							"level1__sibling":         knownvalue.StringExact("true"),
							"top":                     knownvalue.StringExact("simple"),
						}),
					),
				},
			},
		},
	})
}

func TestJsonflattenFunction_RoundTrip(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
				    original = {
						arr = ["x", "y"]
						hello = {
							world = "test"
							foo = "bar"
						}
						top = "level"
					}
				}
				output "roundtrip" {
					value = jsonencode(provider::utils::jsonexpand(provider::utils::jsonflatten(local.original), true))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"roundtrip",
						knownvalue.StringExact(`{"arr":["x","y"],"hello":{"foo":"bar","world":"test"},"top":"level"}`),
					),
				},
			},
		},
	})
}

func TestJsonflattenFunction_RoundTrip_DefaultNoArrayExpansion(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
				    original = {
						arr = ["x", "y"]
						hello = {
							world = "test"
							foo = "bar"
						}
						top = "level"
					}
				}
				output "roundtrip" {
					value = provider::utils::jsonexpand(provider::utils::jsonflatten(local.original))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"roundtrip",
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"arr": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"0": knownvalue.StringExact("x"),
								"1": knownvalue.StringExact("y"),
							}),
							"hello": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"world": knownvalue.StringExact("test"),
								"foo":   knownvalue.StringExact("bar"),
							}),
							"top": knownvalue.StringExact("level"),
						}),
					),
				},
			},
		},
	})
}
