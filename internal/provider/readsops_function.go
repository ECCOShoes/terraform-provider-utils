// Copyright (c) ECCO A/S
// SPDX-License-Identifier: MIT

package provider

import (
	"context"

	"github.com/getsops/sops/v3/decrypt"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = ReadsopsFunction{}
)

func NewReadsopsFunction() function.Function {
	return ReadsopsFunction{}
}

type ReadsopsFunction struct{}

func (r ReadsopsFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "readsops"
}

func (r ReadsopsFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Readsops function",
		MarkdownDescription: "Merges input json into a single json",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "encrypted",
				MarkdownDescription: "Sops-encrypted content",
			},
			function.StringParameter{
				Name:                "format",
				MarkdownDescription: "Can be `json`, `yaml`, `ini`, `dotenv` or `binary`",
			},
		},
		Return: function.StringReturn{},
	}
}

func (r ReadsopsFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var data string
	var format string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &data, &format))

	if resp.Error != nil {
		return
	}
	if len(data) == 0 {
		resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, data))
		return
	}

	result, err := decrypt.Data([]byte(data), format)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, err.Error()))
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, string(result)))
}
