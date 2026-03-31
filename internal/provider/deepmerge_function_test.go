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

func TestDeepmergeFunction_BasicMerge(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(false, {hello={world="test"}}, {hello={qwe="asd"}}))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"hello":{"qwe":"asd","world":"test"}}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_MultipleObjects(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(false, {a=1}, {b=2}, {c=3}))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"a":1,"b":2,"c":3}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_NestedMerge(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						false,
						{nested={deep={value=1}}},
						{nested={deep={another=2}}}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"nested":{"deep":{"another":2,"value":1}}}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_OverwriteTrue(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(true, {key="value1"}, {key="value2"}))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"key":"value2"}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_OverwriteFalse(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(false, {key="value1"}, {key="value2"}))
				}
				`,
				ExpectError: regexp.MustCompile(`(?is)(error due to parameter with value of primitive type|only maps and slices/arrays can be merged|cannot have.*define the same key twice)`),
			},
		},
	})
}

func TestDeepmergeFunction_ComplexMerge(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					obj1 = {
						api = {
							endpoint = "https://example.com"
							version  = "v1"
						}
						features = ["feature1"]
					}
					obj2 = {
						api = {
							timeout = 30
						}
						features = ["feature2"]
					}
				}
				output "test" {
					value = jsonencode(provider::utils::deepmerge(true, local.obj1, local.obj2))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"api":{"endpoint":"https://example.com","timeout":30,"version":"v1"},"features":["feature1","feature2"]}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_WithArrays(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{items=[1, 2, 3]},
						{items=[4, 5, 6]}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"items":[1,2,3,4,5,6]}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_WithNumbers(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						false,
						{count=10, price=19.99},
						{total=100}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"count":10,"price":19.99,"total":100}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_WithBooleans(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						false,
						{enabled=true, debug=false},
						{verbose=true}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"debug":false,"enabled":true,"verbose":true}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_InsufficientObjects(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(false, {a=1}))
				}
				`,
				ExpectError: regexp.MustCompile(`At least 2 objects are required`),
			},
		},
	})
}

func TestDeepmergeFunction_SingleListArgument(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(false, [{a=1}, {b=2}, {c=3}]))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"a":1,"b":2,"c":3}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_SingleListWithNestedMerge(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				locals {
					objects_to_merge = [
						{api = {endpoint = "https://example.com", version = "v1"}},
						{api = {timeout = 30}},
						{features = ["feature1", "feature2"]}
					]
				}
				output "test" {
					value = jsonencode(provider::utils::deepmerge(true, local.objects_to_merge))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"api":{"endpoint":"https://example.com","timeout":30,"version":"v1"},"features":["feature1","feature2"]}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_SingleListInsufficientItems(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(false, [{a=1}]))
				}
				`,
				ExpectError: regexp.MustCompile(`At least 2 objects are required`),
			},
		},
	})
}

func TestDeepmergeFunction_EmptyObjects(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(false, {}, {}))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_MixedTypes(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{
							name   = "example"
							count  = 5
							active = true
							tags   = ["tag1", "tag2"]
							config = {
								setting1 = "value1"
							}
						},
						{
							count  = 10
							tags   = ["tag3"]
							config = {
								setting2 = "value2"
							}
							extra = "data"
						}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"active":true,"config":{"setting1":"value1","setting2":"value2"},"count":10,"extra":"data","name":"example","tags":["tag1","tag2","tag3"]}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_ObjectAndArray(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{data={key="value"}},
						{data=[1, 2, 3]}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"data":[1,2,3]}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_ArrayAndObject(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{data=[1, 2, 3]},
						{data={key="value"}}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"data":{"key":"value"}}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_ScalarAndObject(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{data="simple string"},
						{data={nested="object"}}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"data":{"nested":"object"}}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_ObjectAndScalar(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{data={nested="object"}},
						{data="simple string"}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"data":"simple string"}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_ArrayAndScalar(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{data=[1, 2, 3]},
						{data="replaced"}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"data":"replaced"}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_ScalarAndArray(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{data=123},
						{data=["new", "array"]}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"data":["new","array"]}`),
					),
				},
			},
		},
	})
}

func TestDeepmergeFunction_TypeMismatchWithOverwriteFalse(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						false,
						{data="string"},
						{data=123}
					))
				}
				`,
				ExpectError: regexp.MustCompile(`Failed to merge objects`),
			},
		},
	})
}

func TestDeepmergeFunction_DifferentScalarTypes(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "test" {
					value = jsonencode(provider::utils::deepmerge(
						true,
						{value="string", flag=true, num=42},
						{value=999, flag="yes", num=false}
					))
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue(
						"test",
						knownvalue.StringExact(`{"flag":"yes","num":false,"value":999}`),
					),
				},
			},
		},
	})
}
