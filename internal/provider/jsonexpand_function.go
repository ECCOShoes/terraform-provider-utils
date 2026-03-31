// Copyright (c) ECCO A/S
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ function.Function = JsonexpandFunction{}
)

func NewJsonexpandFunction() function.Function {
	return JsonexpandFunction{}
}

type JsonexpandFunction struct{}

func (r JsonexpandFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "jsonexpand"
}

func (r JsonexpandFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Jsonexpand function",
		MarkdownDescription: "Expands a flat map into a nested object structure using '__' as separator. Optionally converts numeric-key objects into arrays (0, 1, 2, ...) when enabled.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:                "obj",
				MarkdownDescription: "Flat map to expand into nested structure",
			},
		},
		VariadicParameter: function.BoolParameter{
			Name:                "expand_arrays",
			MarkdownDescription: "Optional. If true, objects whose keys are all numeric indices are expanded into arrays (missing indices become null). Default is false.",
		},
		Return: function.DynamicReturn{},
	}
}

func convertNumericKeyObjectsToTuples(ctx context.Context, v attr.Value, enable bool) (attr.Value, error) {
	if !enable {
		return v, nil
	}

	obj, ok := v.(types.Object)
	if !ok {
		return v, nil
	}

	attrs := obj.Attributes()
	converted := make(map[string]attr.Value, len(attrs))
	for k, child := range attrs {
		cv, err := convertNumericKeyObjectsToTuples(ctx, child, enable)
		if err != nil {
			return nil, err
		}
		converted[k] = cv
	}

	if len(converted) == 0 {
		// Empty object stays an object.
		return obj, nil
	}

	maxIdx := -1
	indices := make(map[int]attr.Value, len(converted))
	for k, child := range converted {
		idx, err := strconv.Atoi(k)
		if err != nil || idx < 0 {
			// Not a pure numeric-key object: rebuild it as an object with converted children.
			attrTypes := make(map[string]attr.Type, len(converted))
			for kk, vv := range converted {
				attrTypes[kk] = vv.Type(ctx)
			}
			newObj, diags := types.ObjectValue(attrTypes, converted)
			if diags.HasError() {
				return nil, fmt.Errorf("error creating nested object: %s", diags)
			}
			return newObj, nil
		}
		indices[idx] = child
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	// All keys are numeric indices: build a tuple from 0..max.
	elems := make([]attr.Value, maxIdx+1)
	elemTypes := make([]attr.Type, maxIdx+1)
	for i := 0; i <= maxIdx; i++ {
		if child, ok := indices[i]; ok {
			elems[i] = child
			elemTypes[i] = child.Type(ctx)
		} else {
			nullDyn := types.DynamicNull()
			elems[i] = nullDyn
			elemTypes[i] = nullDyn.Type(ctx)
		}
	}

	tupleVal, diags := types.TupleValue(elemTypes, elems)
	if diags.HasError() {
		return nil, fmt.Errorf("error creating tuple value: %s", diags)
	}
	return tupleVal, nil
}

func expandMap(flat map[string]attr.Value, sep string) (map[string]attr.Value, error) {
	root := make(map[string]attr.Value)

	for flatKey, value := range flat {
		parts := strings.Split(flatKey, sep)
		if len(parts) == 1 {
			// No separator, just a simple key
			root[flatKey] = value
			continue
		}

		// Build nested structure
		err := setNestedValue(root, parts, value)
		if err != nil {
			return nil, err
		}
	}

	return root, nil
}

func setNestedValue(root map[string]attr.Value, parts []string, value attr.Value) error {
	if len(parts) == 0 {
		return fmt.Errorf("empty key parts")
	}

	if len(parts) == 1 {
		root[parts[0]] = value
		return nil
	}

	// Get or create intermediate map
	firstKey := parts[0]
	var nextMap map[string]attr.Value

	if existing, ok := root[firstKey]; ok {
		if existingObj, isObj := existing.(types.Object); isObj {
			nextMap = existingObj.Attributes()
		} else {
			return fmt.Errorf("conflicting key: %s is already set as a scalar value", firstKey)
		}
	} else {
		nextMap = make(map[string]attr.Value)
	}

	// Recursively set the value
	err := setNestedValue(nextMap, parts[1:], value)
	if err != nil {
		return err
	}

	// Create object from the map
	attrTypes := make(map[string]attr.Type)
	for k, v := range nextMap {
		attrTypes[k] = v.Type(context.Background())
	}

	obj, diags := types.ObjectValue(attrTypes, nextMap)
	if diags.HasError() {
		return fmt.Errorf("error creating nested object: %s", diags)
	}

	root[firstKey] = obj
	return nil
}

func (r JsonexpandFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var data types.Dynamic
	var expandArraysArgs []bool

	resp.Error = function.ConcatFuncErrors(req.Arguments.GetArgument(ctx, 0, &data))

	if resp.Error != nil {
		return
	}

	// Optional expand_arrays flag (variadic, 0 or 1 values)
	resp.Error = function.ConcatFuncErrors(req.Arguments.GetArgument(ctx, 1, &expandArraysArgs))
	if resp.Error != nil {
		return
	}
	if len(expandArraysArgs) > 1 {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(1, "expand_arrays accepts at most one boolean"))
		return
	}
	expandArrays := false
	if len(expandArraysArgs) == 1 {
		expandArrays = expandArraysArgs[0]
	}

	// Extract the underlying value
	underlyingValue := data.UnderlyingValue()

	var stringMap map[string]attr.Value

	// Check if it's a map or object
	if inputMap, ok := underlyingValue.(types.Map); ok {
		// Get the map elements
		mapElements := inputMap.Elements()

		// Validate all values are strings
		stringMap = make(map[string]attr.Value)
		for k, v := range mapElements {
			if _, ok := v.(types.String); !ok {
				resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, fmt.Sprintf("All map values must be strings, but key '%s' has type %T", k, v)))
				return
			}
			stringMap[k] = v
		}
	} else if inputObj, ok := underlyingValue.(types.Object); ok {
		// Get the object attributes
		attrMap := inputObj.Attributes()

		// Validate all values are strings
		stringMap = make(map[string]attr.Value)
		for k, v := range attrMap {
			if _, ok := v.(types.String); !ok {
				resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, fmt.Sprintf("All object values must be strings, but key '%s' has type %T", k, v)))
				return
			}
			stringMap[k] = v
		}
	} else {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, "Input must be a map/object. Arrays and scalars are not supported."))
		return
	}

	// Expand the map
	result, err := expandMap(stringMap, "__")
	if err != nil {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, fmt.Sprintf("Error expanding map: %s", err)))
		return
	}

	// Determine result type dynamically
	attrTypes := make(map[string]attr.Type)
	for k, v := range result {
		attrTypes[k] = v.Type(ctx)
	}

	// Create final object
	resultObj, diags := types.ObjectValue(attrTypes, result)
	if diags.HasError() {
		resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, fmt.Sprintf("Error creating result object: %s", diags)))
		return
	}

	finalVal := attr.Value(resultObj)
	if expandArrays {
		converted, convErr := convertNumericKeyObjectsToTuples(ctx, resultObj, true)
		if convErr != nil {
			resp.Error = function.ConcatFuncErrors(function.NewArgumentFuncError(0, fmt.Sprintf("Error converting numeric-key objects to arrays: %s", convErr)))
			return
		}
		finalVal = converted
	}

	// Wrap in dynamic for return
	resultDynamic := types.DynamicValue(finalVal)
	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, resultDynamic))
}
