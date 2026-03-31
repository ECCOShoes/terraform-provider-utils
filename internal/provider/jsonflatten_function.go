// Copyright (c) ECCO A/S
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ function.Function = JsonflattenFunction{}
)

func NewJsonflattenFunction() function.Function {
	return JsonflattenFunction{}
}

type JsonflattenFunction struct{}

func (r JsonflattenFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "jsonflatten"
}

func (r JsonflattenFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Jsonflatten function",
		MarkdownDescription: "Flattens a nested object/map (and any nested arrays) into a single-level map with keys separated by '__'. Arrays are flattened using numeric indices (0, 1, 2, ...).",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:                "obj",
				MarkdownDescription: "Object/map (or array) to flatten",
			},
		},
		Return: function.MapReturn{
			ElementType: types.StringType,
		},
	}
}

func flattenValue(root interface{}, sep string) (map[string]attr.Value, error) {
	result := make(map[string]attr.Value)
	if err := flattenRecursive("", root, sep, result); err != nil {
		return nil, err
	}
	return result, nil
}

func flattenRecursive(prefix string, current interface{}, sep string, result map[string]attr.Value) error {
	switch curr := current.(type) {
	case map[string]attr.Value:
		for k, v := range curr {
			fullKey := k
			if prefix != "" {
				fullKey = prefix + sep + k
			}
			err := flattenRecursive(fullKey, v, sep, result)
			if err != nil {
				return err
			}
		}
		return nil

	case basetypes.ObjectValuable:
		objVal, diags := curr.ToObjectValue(context.Background())
		if diags.HasError() {
			return fmt.Errorf("failed to read object value: %s", diags)
		}
		if objVal.IsNull() || objVal.IsUnknown() {
			return fmt.Errorf("null/unknown object value")
		}
		attrMap := objVal.Attributes()
		for k, v := range attrMap {
			fullKey := k
			if prefix != "" {
				fullKey = prefix + sep + k
			}
			if err := flattenRecursive(fullKey, v, sep, result); err != nil {
				return err
			}
		}
		return nil

	case basetypes.MapValuable:
		mapVal, diags := curr.ToMapValue(context.Background())
		if diags.HasError() {
			return fmt.Errorf("failed to read map value: %s", diags)
		}
		if mapVal.IsNull() || mapVal.IsUnknown() {
			return fmt.Errorf("null/unknown map value")
		}
		elems := mapVal.Elements()
		for k, v := range elems {
			fullKey := k
			if prefix != "" {
				fullKey = prefix + sep + k
			}
			if err := flattenRecursive(fullKey, v, sep, result); err != nil {
				return err
			}
		}
		return nil

	case types.String:
		result[prefix] = curr
		return nil

	case types.Number:
		result[prefix] = types.StringValue(curr.String())
		return nil

	case types.Bool:
		result[prefix] = types.StringValue(curr.String())
		return nil

	case basetypes.ListValuable:
		listVal, diags := curr.ToListValue(context.Background())
		if diags.HasError() {
			return fmt.Errorf("failed to read list value: %s", diags)
		}
		if listVal.IsNull() || listVal.IsUnknown() {
			return fmt.Errorf("null/unknown list value")
		}
		elems := listVal.Elements()
		for i, v := range elems {
			idx := strconv.Itoa(i)
			fullKey := idx
			if prefix != "" {
				fullKey = prefix + sep + idx
			}
			if err := flattenRecursive(fullKey, v, sep, result); err != nil {
				return err
			}
		}
		return nil

	case basetypes.TupleValue:
		if curr.IsNull() || curr.IsUnknown() {
			return fmt.Errorf("null/unknown tuple value")
		}
		elems := curr.Elements()
		for i, v := range elems {
			idx := strconv.Itoa(i)
			fullKey := idx
			if prefix != "" {
				fullKey = prefix + sep + idx
			}
			if err := flattenRecursive(fullKey, v, sep, result); err != nil {
				return err
			}
		}
		return nil

	case basetypes.SetValuable:
		setVal, diags := curr.ToSetValue(context.Background())
		if diags.HasError() {
			return fmt.Errorf("failed to read set value: %s", diags)
		}
		if setVal.IsNull() || setVal.IsUnknown() {
			return fmt.Errorf("null/unknown set value")
		}
		elems := setVal.Elements()
		for i, v := range elems {
			idx := strconv.Itoa(i)
			fullKey := idx
			if prefix != "" {
				fullKey = prefix + sep + idx
			}
			if err := flattenRecursive(fullKey, v, sep, result); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("unsupported type: %T", curr)
	}
}

func (r JsonflattenFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var data types.Dynamic

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &data))

	if resp.Error != nil {
		return
	}

	// Extract the underlying value
	underlyingValue := data.UnderlyingValue()

	// Only accept objects/maps or arrays at the root
	switch underlyingValue.(type) {
	case basetypes.ObjectValuable, basetypes.MapValuable, basetypes.ListValuable, basetypes.SetValuable, basetypes.TupleValue:
		// ok
	default:
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, "Input must be a map/object or array. Scalars are not supported."))
		return
	}

	result, err := flattenValue(underlyingValue, "__")
	if err != nil {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, fmt.Sprintf("Error flattening value: %s", err)))
		return
	}

	// Convert to types.Map
	resultMap, diags := types.MapValue(types.StringType, result)
	if diags.HasError() {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, fmt.Sprintf("Error creating result map: %s", diags)))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, resultMap))
}
