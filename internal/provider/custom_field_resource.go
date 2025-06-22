// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0
package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-piano/internal/piano_id"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = &CustomFieldResource{}
)

type CustomFieldResource struct {
	client *piano_id.Client
}

func NewCustomFieldResource() resource.Resource {
	return &CustomFieldResource{}
}
func (r *CustomFieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(PianoProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected PianoProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = &client.idClient
}

func (r *CustomFieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unsafe_custom_field"
}

type CustomFieldResourceModel struct {
	Aid               types.String    `tfsdk:"aid"` // The application ID
	FieldName         types.String    `tfsdk:"field_name"`
	Title             types.String    `tfsdk:"title"`
	Comment           types.String    `tfsdk:"comment"`
	Editable          types.Bool      `tfsdk:"editable"`
	DataType          types.String    `tfsdk:"data_type"`
	Options           *[]types.String `tfsdk:"options"`
	RequiredByDefault types.Bool      `tfsdk:"required_by_default"`
	Archived          types.Bool      `tfsdk:"archived"`
	DefaultSortOrder  types.Int32     `tfsdk:"default_sort_order"`
	DefaultValue      types.String    `tfsdk:"default_value"`
	Multiline         types.Bool      `tfsdk:"multiline"`
	Validators        *[]types.String `tfsdk:"validators"`
}

func (*CustomFieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This is a custom field resource. This resource is unsafe in that it always creates or updates resources" +
			" because piano id API does not provide a way of getting custom field without mutating it",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"field_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Piano ID custom field name, which serves as an identifier for custom field",
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Piano ID custom field title(friendly name)",
			},
			"comment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Piano ID custom field internal comment",
			},
			"editable": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Piano ID custom field editability",
			},
			"data_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"BOOLEAN", "TEXT", "ISO_DATE", "SINGLE_SELECT_LIST", "MULTI_SELECT_LIST",
					),
				},
				MarkdownDescription: "Piano ID custom field type",
			},
			"options": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Piano ID custom field select options",
			},
			"required_by_default": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Piano ID custom field archive status(default: false)",
			},
			"archived": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Piano ID custom field archive status(default: false)",
			},
			"default_sort_order": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "Piano ID custom field default sort order",
			},
			"default_value": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Piano ID custom field default value",
			},
			"multiline": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Piano ID custom field multiline setting for TEXT data type",
			},
			"validators": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Piano ID custom field validators",
			},
		},
	}
}

func (r *CustomFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Warn(ctx, "Read operation is not supported for custom field resource as piano id exposes only create/update API")
	tflog.Debug(ctx, "To get custom fields, send a request to *.piano.io/api/v3/publisher/user/get endpoint")
}
func (r *CustomFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state CustomFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	options := []string{}
	if state.Options != nil {
		for _, option := range *state.Options {
			options = append(options, option.ValueString())
		}
	}
	validators := []string{}
	if state.Validators != nil {
		for _, validator := range *state.Validators {
			validators = append(validators, validator.ValueString())
		}
	}
	request := piano_id.CustomFieldDefinition{
		FieldName:         state.FieldName.ValueString(),
		Title:             state.Title.ValueString(),
		Comment:           state.Comment.ValueStringPointer(),
		Editable:          state.Editable.ValueBool(),
		DataType:          piano_id.CustomFieldDefinitionDataType(state.DataType.ValueString()),
		Options:           options,
		RequiredByDefault: state.RequiredByDefault.ValueBool(),
		Archived:          false,
		DefaultSortOrder:  state.DefaultSortOrder.ValueInt32Pointer(),
		Attribute: piano_id.CustomFieldAttribute{
			DefaultValue: state.DefaultValue.ValueStringPointer(),
			Multiline:    state.Multiline.ValueBoolPointer(),
		},
		Validators: validators,
		Tooltip:    &piano_id.Tooltip{},
	}
	tflog.Info(ctx, fmt.Sprintf("creating custom_field: %s of type %s", state.FieldName.ValueString(), state.DataType.ValueString()))
	response, err := r.client.PublisherCustomFieldPost(ctx, []piano_id.CustomFieldDefinition{
		request,
	})
	if err != nil {
		tflog.Info(ctx, fmt.Sprintf("unable to create custom_field due to %e", err))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create or update custom field, got error: %s", err))
		return
	}
	result, err := piano_id.ParsePublisherCustomFieldPostResponse(response)
	if err != nil {
		tflog.Info(ctx, fmt.Sprintf("unable to parse custom_field response due to %e", err))
		resp.Diagnostics.AddError("Marshal Error", fmt.Sprintf("Unable to parse response as ParsePublisherCustomFieldPostResponse, got error: %s", err))
		return
	}

	var data piano_id.CustomFieldDefinition
	if result.JSON200 != nil {
		if len(*result.JSON200) > 0 {
			data = (*result.JSON200)[0]
		} else {
			resp.Diagnostics.AddError("Invalid State", "Piano ID API returned empty response for non empty request")
			return
		}
	} else {
		messages := []string{}
		for _, message := range result.JSONDefault.ErrorCodeList {
			messages = append(messages, message.Message)
		}
		tflog.Error(ctx, "errors: ["+strings.Join(messages, ",")+"]")
		resp.Diagnostics.AddError("Status Error", fmt.Sprintf("Unable to create custom field due to %s", "["+strings.Join(messages, ",")+"]"))
		return
	}

	state.FieldName = types.StringValue(data.FieldName)
	state.Title = types.StringValue(data.Title)
	state.Comment = types.StringPointerValue(data.Comment)
	state.Editable = types.BoolValue(data.Editable)
	state.DataType = types.StringValue(string(data.DataType))
	if state.Options != nil {
		optionsFromResponse := []types.String{}
		for _, option := range data.Options {
			optionsFromResponse = append(optionsFromResponse, types.StringValue(option))
		}
		state.Options = &optionsFromResponse
	}
	state.RequiredByDefault = types.BoolValue(data.RequiredByDefault)
	state.DefaultValue = types.StringPointerValue(data.Attribute.DefaultValue)
	state.Multiline = types.BoolPointerValue(data.Attribute.Multiline)
	state.Archived = types.BoolValue(data.Archived)
	if state.Validators != nil {
		validatorsFromResponse := []types.String{}
		for _, validator := range data.Validators {
			validatorsFromResponse = append(validatorsFromResponse, types.StringValue(validator))
		}
		state.Validators = &validatorsFromResponse
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *CustomFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state CustomFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	options := []string{}
	if state.Options != nil {
		for _, option := range *state.Options {
			options = append(options, option.ValueString())
		}
	}
	validators := []string{}
	if state.Validators != nil {
		for _, validator := range *state.Validators {
			validators = append(validators, validator.ValueString())
		}
	}

	response, err := r.client.PublisherCustomFieldPost(ctx, []piano_id.CustomFieldDefinition{
		{
			FieldName:         state.FieldName.ValueString(),
			Title:             state.Title.ValueString(),
			Comment:           state.Comment.ValueStringPointer(),
			Editable:          state.Editable.ValueBool(),
			DataType:          piano_id.CustomFieldDefinitionDataType(state.DataType.ValueString()),
			Options:           options,
			RequiredByDefault: state.RequiredByDefault.ValueBool(),
			Archived:          false,
			DefaultSortOrder:  state.DefaultSortOrder.ValueInt32Pointer(),
			Attribute: piano_id.CustomFieldAttribute{
				DefaultValue: state.DefaultValue.ValueStringPointer(),
				Multiline:    state.Multiline.ValueBoolPointer(),
			},
			Validators: validators,
			Tooltip:    &piano_id.Tooltip{},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create or update custom field, got error: %s", err))
		return
	}
	result, err := piano_id.ParsePublisherCustomFieldPostResponse(response)
	if err != nil {
		resp.Diagnostics.AddError("Marshal Error", fmt.Sprintf("Unable to parse response as ParsePublisherCustomFieldPostResponse, got error: %s", err))
		return
	}
	var data piano_id.CustomFieldDefinition
	if len(*result.JSON200) > 0 {
		data = (*result.JSON200)[0]
	} else {
		resp.Diagnostics.AddError("Invalid State", "Piano ID API returned empty response for non empty request")
		return
	}
	state.FieldName = types.StringValue(data.FieldName)
	state.Title = types.StringValue(data.Title)
	state.Comment = types.StringPointerValue(data.Comment)
	state.Editable = types.BoolValue(data.Editable)
	state.DataType = types.StringValue(string(data.DataType))
	if state.Options != nil {
		optionsFromResponse := []types.String{}
		for _, option := range data.Options {
			optionsFromResponse = append(optionsFromResponse, types.StringValue(option))
		}
		state.Options = &optionsFromResponse
	}
	state.RequiredByDefault = types.BoolValue(data.RequiredByDefault)
	state.Archived = types.BoolValue(data.Archived)
	state.DefaultValue = types.StringPointerValue(data.Attribute.DefaultValue)
	state.Multiline = types.BoolPointerValue(data.Attribute.Multiline)
	if state.Validators != nil {
		validatorsFromResponse := []types.String{}
		for _, validator := range data.Validators {
			validatorsFromResponse = append(validatorsFromResponse, types.StringValue(validator))
		}
		state.Validators = &validatorsFromResponse
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *CustomFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	options := []string{}
	if state.Options != nil {
		for _, option := range *state.Options {
			options = append(options, option.ValueString())
		}
	}
	validators := []string{}
	if state.Validators != nil {
		for _, validator := range *state.Validators {
			validators = append(validators, validator.ValueString())
		}
	}

	tflog.Warn(ctx, fmt.Sprintf("(logically) deleting custom field by setting archived=true for %s", state.FieldName.ValueString()))
	response, err := r.client.PublisherCustomFieldPost(ctx, []piano_id.CustomFieldDefinition{
		{
			FieldName:         state.FieldName.ValueString(),
			Title:             state.Title.ValueString(),
			Comment:           state.Comment.ValueStringPointer(),
			Editable:          state.Editable.ValueBool(),
			DataType:          piano_id.CustomFieldDefinitionDataType(state.DataType.ValueString()),
			Options:           options,
			RequiredByDefault: state.RequiredByDefault.ValueBool(),
			Archived:          true,
			DefaultSortOrder:  state.DefaultSortOrder.ValueInt32Pointer(),
			Attribute: piano_id.CustomFieldAttribute{
				DefaultValue: state.DefaultValue.ValueStringPointer(),
				Multiline:    state.Multiline.ValueBoolPointer(),
			},
			Validators: validators,
			Tooltip:    &piano_id.Tooltip{},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update custom field, got error: %s", err))
		return
	}
	result, err := piano_id.ParsePublisherCustomFieldPostResponse(response)
	if err != nil {
		resp.Diagnostics.AddError("Marshal Error", fmt.Sprintf("Unable to parse response as ParsePublisherCustomFieldPostResponse, got error: %s", err))
		return
	}
	var data piano_id.CustomFieldDefinition
	if len(*result.JSON200) > 0 {
		data = (*result.JSON200)[0]
	} else {
		resp.Diagnostics.AddError("Invalid State", "Piano ID API returned empty response for non empty request")
		return
	}
	if !data.Archived {
		resp.Diagnostics.AddError("Invalid State", "Piano ID API returned `archived=false` for deleted resource")
		return
	}
}
