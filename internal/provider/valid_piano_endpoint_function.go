// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"terraform-provider-piano/internal/piano"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = ValidPianoEndpoint{}
)

func NewValidPianoEndpoint() function.Function {
	return ValidPianoEndpoint{}
}

type ValidPianoEndpoint struct{}

func (r ValidPianoEndpoint) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "valid_piano_endpoint"
}

func (r ValidPianoEndpoint) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "valid_piano_endpoint function checks if a string starts with piano.io api endpoint endpoint",
		MarkdownDescription: "Returns true if the given string starts with one of the predefined piano.io api endpoint",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "input",
				MarkdownDescription: "Arbitrary string to be validated",
			},
		},
		Return: function.BoolReturn{},
	}
}

func (r ValidPianoEndpoint) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var data string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &data))

	if resp.Error != nil {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, piano.ValidPianoEndpoint(data)))
}
