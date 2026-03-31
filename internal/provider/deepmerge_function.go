// Copyright (c) ECCO A/S
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ function.Function = DeepmergeFunction{}
)

func NewDeepmergeFunction() function.Function {
	return DeepmergeFunction{}
}

type DeepmergeFunction struct{}

func (r DeepmergeFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "deepmerge"
}

func (r DeepmergeFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Deepmerge function",
		MarkdownDescription: `Merges multiple objects into a single dynamic value.

Merge behavior:
- **Objects**: Merged recursively (nested objects are combined)
- **Arrays**: Concatenated (all elements from all arrays are combined)
- **Scalars** (strings, numbers, booleans): Second value overwrites the first when overwrite=true

When overwrite=false, any conflicting keys with primitive values will cause an error.
When overwrite=true, conflicts are resolved by merging intelligently based on the value types.

Special behavior: If a single list/array argument is provided, all items in that list will be merged together.`,
		VariadicParameter: function.DynamicParameter{
			Name:                "objects",
			MarkdownDescription: "Objects to merge (at least 2 required), or a single list containing objects to merge",
		},
		Parameters: []function.Parameter{
			function.BoolParameter{
				Name:                "overwrite",
				MarkdownDescription: "Allow overwriting same primitive keys. Set to false to error on conflicts, true to merge intelligently.",
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (r DeepmergeFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var overwrite bool
	var objects []basetypes.DynamicValue

	// Get the overwrite parameter (first positional parameter)
	resp.Error = function.ConcatFuncErrors(req.Arguments.GetArgument(ctx, 0, &overwrite))
	if resp.Error != nil {
		return
	}

	// Get variadic objects
	resp.Error = function.ConcatFuncErrors(req.Arguments.GetArgument(ctx, 1, &objects))
	if resp.Error != nil {
		return
	}

	// Special case: if only one argument is provided, check if it's a list of objects to merge
	if len(objects) == 1 {
		singleArg, err := dynamicToGo(objects[0])
		if err != nil {
			resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "Failed to convert argument: "+err.Error()))
			return
		}

		// If it's a slice/list, treat it as a list of objects to merge
		if slice, isSlice := singleArg.([]interface{}); isSlice {
			if len(slice) < 2 {
				resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "At least 2 objects are required for merging"))
				return
			}
			// Merge all items in the list
			result := slice[0]
			for i := 1; i < len(slice); i++ {
				result, err = mergeValues(result, slice[i], overwrite)
				if err != nil {
					resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "Failed to merge objects: "+err.Error()))
					return
				}
			}

			// Convert result back to Terraform dynamic value
			dynamicVal, err := goToTerraformDynamic(ctx, result)
			if err != nil {
				resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "Failed to convert result to dynamic value: "+err.Error()))
				return
			}

			resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, dynamicVal))
			return
		}

		// Not a list, require at least 2 objects
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "At least 2 objects are required for merging"))
		return
	}

	// Convert first object to Go value
	result, err := dynamicToGo(objects[0])
	if err != nil {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "Failed to convert first object: "+err.Error()))
		return
	}

	// Merge remaining objects
	for i := 1; i < len(objects); i++ {
		current, err := dynamicToGo(objects[i])
		if err != nil {
			resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "Failed to convert object: "+err.Error()))
			return
		}

		result, err = mergeValues(result, current, overwrite)
		if err != nil {
			resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "Failed to merge objects: "+err.Error()))
			return
		}
	}

	// Convert result back to Terraform dynamic value
	dynamicVal, err := goToTerraformDynamic(ctx, result)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "Failed to convert result to dynamic value: "+err.Error()))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, dynamicVal))
}

// mergeValues performs a deep merge of two arbitrary Go values
func mergeValues(a, b interface{}, overwrite bool) (interface{}, error) {
	// If types are different, decide based on overwrite flag
	aMap, aIsMap := a.(map[string]interface{})
	bMap, bIsMap := b.(map[string]interface{})
	aSlice, aIsSlice := a.([]interface{})
	bSlice, bIsSlice := b.([]interface{})

	// Case 1: Both are maps - merge recursively
	if aIsMap && bIsMap {
		result := make(map[string]interface{})
		// Copy all keys from a
		for k, v := range aMap {
			result[k] = v
		}
		// Merge keys from b
		for k, v := range bMap {
			if existingVal, exists := result[k]; exists {
				// Key exists in both - recursively merge
				merged, err := mergeValues(existingVal, v, overwrite)
				if err != nil {
					return nil, err
				}
				result[k] = merged
			} else {
				// Key only in b
				result[k] = v
			}
		}
		return result, nil
	}

	// Case 2: Both are slices/arrays - concatenate
	if aIsSlice && bIsSlice {
		result := make([]interface{}, 0, len(aSlice)+len(bSlice))
		result = append(result, aSlice...)
		result = append(result, bSlice...)
		return result, nil
	}

	// Case 3: Types don't match or both are scalars
	// If types differ or both are scalars
	if overwrite {
		// Use second value
		return b, nil
	} else {
		// Error on conflict
		return nil, fmt.Errorf("cannot have parameters define the same key twice")
	}
}

// dynamicToGo converts a DynamicValue to a Go value (map, slice, or scalar)
func dynamicToGo(dv basetypes.DynamicValue) (interface{}, error) {
	// Get the underlying value and convert to Go
	underlyingValue := dv.UnderlyingValue()
	return terraformValueToGo(underlyingValue)
}

// goToTerraformDynamic converts a Go value back to a Terraform DynamicValue
func goToTerraformDynamic(ctx context.Context, val interface{}) (basetypes.DynamicValue, error) {
	switch v := val.(type) {
	case map[string]interface{}:
		// Convert map to object
		attrTypes := make(map[string]attr.Type)
		attrValues := make(map[string]attr.Value)
		for key, value := range v {
			dynVal, err := goToTerraformDynamic(ctx, value)
			if err != nil {
				return basetypes.NewDynamicNull(), err
			}
			attrTypes[key] = types.DynamicType
			attrValues[key] = dynVal
		}
		objVal, diags := types.ObjectValue(attrTypes, attrValues)
		if diags.HasError() {
			return basetypes.NewDynamicNull(), fmt.Errorf("failed to create object value: %v", diags)
		}
		return types.DynamicValue(objVal), nil

	case []interface{}:
		// Convert slice to tuple
		elems := make([]attr.Value, 0, len(v))
		elemTypes := make([]attr.Type, 0, len(v))
		for _, item := range v {
			dynVal, err := goToTerraformDynamic(ctx, item)
			if err != nil {
				return basetypes.NewDynamicNull(), err
			}
			elems = append(elems, dynVal)
			elemTypes = append(elemTypes, types.DynamicType)
		}
		tupleVal, diags := types.TupleValue(elemTypes, elems)
		if diags.HasError() {
			return basetypes.NewDynamicNull(), fmt.Errorf("failed to create tuple value: %v", diags)
		}
		return types.DynamicValue(tupleVal), nil

	case string:
		return types.DynamicValue(types.StringValue(v)), nil

	case float64:
		numVal := types.NumberValue(new(big.Float).SetFloat64(v))
		return types.DynamicValue(numVal), nil

	case bool:
		return types.DynamicValue(types.BoolValue(v)), nil

	case nil:
		return types.DynamicNull(), nil

	default:
		return basetypes.NewDynamicNull(), fmt.Errorf("unsupported type: %T", v)
	}
}

// terraformValueToGo recursively converts Terraform values to Go native types
func terraformValueToGo(val interface{}) (interface{}, error) {
	switch v := val.(type) {
	case basetypes.StringValuable:
		strVal, _ := v.ToStringValue(context.Background())
		if strVal.IsNull() || strVal.IsUnknown() {
			return nil, nil
		}
		return strVal.ValueString(), nil
	case basetypes.NumberValuable:
		numVal, _ := v.ToNumberValue(context.Background())
		if numVal.IsNull() || numVal.IsUnknown() {
			return nil, nil
		}
		f, _ := numVal.ValueBigFloat().Float64()
		return f, nil
	case basetypes.BoolValuable:
		boolVal, _ := v.ToBoolValue(context.Background())
		if boolVal.IsNull() || boolVal.IsUnknown() {
			return nil, nil
		}
		return boolVal.ValueBool(), nil
	case basetypes.ListValuable:
		listVal, _ := v.ToListValue(context.Background())
		if listVal.IsNull() || listVal.IsUnknown() {
			return nil, nil
		}
		elems := listVal.Elements()
		result := make([]interface{}, 0, len(elems))
		for _, elem := range elems {
			converted, err := terraformValueToGo(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
		return result, nil
	case basetypes.TupleValue:
		if v.IsNull() || v.IsUnknown() {
			return nil, nil
		}
		elems := v.Elements()
		result := make([]interface{}, 0, len(elems))
		for _, elem := range elems {
			converted, err := terraformValueToGo(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
		return result, nil
	case basetypes.SetValuable:
		setVal, _ := v.ToSetValue(context.Background())
		if setVal.IsNull() || setVal.IsUnknown() {
			return nil, nil
		}
		elems := setVal.Elements()
		result := make([]interface{}, 0, len(elems))
		for _, elem := range elems {
			converted, err := terraformValueToGo(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
		return result, nil
	case basetypes.MapValuable:
		mapVal, _ := v.ToMapValue(context.Background())
		if mapVal.IsNull() || mapVal.IsUnknown() {
			return nil, nil
		}
		elems := mapVal.Elements()
		m := make(map[string]interface{})
		for k, elemVal := range elems {
			converted, err := terraformValueToGo(elemVal)
			if err != nil {
				return nil, err
			}
			m[k] = converted
		}
		return m, nil
	case basetypes.ObjectValuable:
		objVal, _ := v.ToObjectValue(context.Background())
		if objVal.IsNull() || objVal.IsUnknown() {
			return nil, nil
		}
		attrs := objVal.Attributes()
		m := make(map[string]interface{})
		for k, attrVal := range attrs {
			converted, err := terraformValueToGo(attrVal)
			if err != nil {
				return nil, err
			}
			m[k] = converted
		}
		return m, nil
	default:
		return nil, nil
	}
}
