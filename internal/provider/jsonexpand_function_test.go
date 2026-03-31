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

func TestJsonExpandFunction_Known(t *testing.T) {
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
						hello__world = "testvalue"
					}
				}
				output "test" {
					value = provider::utils::jsonexpand(local.tst)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"hello": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"world": knownvalue.StringExact("testvalue"),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestJsonExpandFunction_Null(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::utils::jsonexpand(null)
				}
				`,
				// The parameter does not enable AllowNullValue
				ExpectError: regexp.MustCompile(`argument must not be null`),
			},
		},
	})
}

func TestJsonExpandFunction_Array(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::utils::jsonexpand(["test", "value"])
				}
				`,
				ExpectError: regexp.MustCompile(`Input must be a map/object`),
			},
		},
	})
}

func TestJsonExpandFunction_Scalar(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = provider::utils::jsonexpand("testvalue")
				}
				`,
				ExpectError: regexp.MustCompile(`Input must be a map/object`),
			},
		},
	})
}

func TestJsonExpandFunction_MultipleKeys(t *testing.T) {
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
						hello__world = "test"
						hello__asd = "45"
					}
				}
				output "test" {
					value = provider::utils::jsonexpand(local.tst)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"hello": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"world": knownvalue.StringExact("test"),
								"asd":   knownvalue.StringExact("45"),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestJsonExpandFunction_ArrayExpansion_DefaultDisabled(t *testing.T) {
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
						hello__0 = "d"
						hello__1 = "e"
					}
				}
				output "test" {
					value = provider::utils::jsonexpand(local.tst)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"hello": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"0": knownvalue.StringExact("d"),
								"1": knownvalue.StringExact("e"),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestJsonExpandFunction_ArrayExpansion_EnabledNested(t *testing.T) {
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
						hello__0 = "d"
						hello__1 = "e"
					}
				}
				output "test" {
					value = jsonencode(provider::utils::jsonexpand(local.tst, true))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"hello":["d","e"]}`),
					),
				},
			},
		},
	})
}

func TestJsonExpandFunction_ArrayExpansion_EnabledMissingIndices(t *testing.T) {
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
						"0" = "a"
						"2" = "c"
					}
				}
				output "test" {
					value = jsonencode(provider::utils::jsonexpand(local.tst, true))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`["a",null,"c"]`),
					),
				},
			},
		},
	})
}

func TestJsonExpandFunction_Complex(t *testing.T) {
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
						level1__level2__level3 = "deep"
						level1__level2__another = "value"
						level1__sibling = "test"
						top = "simple"
					}
				}
				output "test" {
					value = provider::utils::jsonexpand(local.tst)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"level1": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"level2": knownvalue.ObjectExact(map[string]knownvalue.Check{
									"level3":  knownvalue.StringExact("deep"),
									"another": knownvalue.StringExact("value"),
								}),
								"sibling": knownvalue.StringExact("test"),
							}),
							"top": knownvalue.StringExact("simple"),
						}),
					),
				},
			},
		},
	})
}
