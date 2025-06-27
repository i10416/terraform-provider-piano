// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0
package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-piano/internal/piano_id"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	Aid                  types.String           `tfsdk:"aid"` // The application ID
	FieldName            types.String           `tfsdk:"field_name"`
	Title                types.String           `tfsdk:"title"`
	Comment              types.String           `tfsdk:"comment"`
	Editable             types.Bool             `tfsdk:"editable"`
	DataType             types.String           `tfsdk:"data_type"`
	Options              *[]types.String        `tfsdk:"options"`
	RequiredByDefault    types.Bool             `tfsdk:"required_by_default"`
	Archived             types.Bool             `tfsdk:"archived"`
	DefaultSortOrder     types.Int32            `tfsdk:"default_sort_order"`
	DefaultValue         types.String           `tfsdk:"default_value"`
	Prechecked           types.Bool             `tfsdk:"prechecked"`
	Placeholder          types.String           `tfsdk:"placeholder"`
	DateFormat           types.String           `tfsdk:"date_format"`
	Global               types.Bool             `tfsdk:"global"`
	Multiline            types.Bool             `tfsdk:"multiline"`
	PreSelectCountryById types.Bool             `tfsdk:"pre_select_country_by_ip"`
	Validators           *[]types.String        `tfsdk:"validators"`
	LengthValidator      *StringLengthValidator `tfsdk:"length_validator"`
	RegexValidator       *RegexValidator        `tfsdk:"regex_validator"`
	EmailValidator       *EmailValidator        `tfsdk:"email_validator"`
	AllowListValidator   *AllowListValidator    `tfsdk:"allow_list_validator"`
	DenyListValidator    *DenyListValidator     `tfsdk:"deny_list_validator"`
}

type StringLengthValidator struct {
	MinLength    types.Int32  `tfsdk:"min_length"`
	MaxLength    types.Int32  `tfsdk:"max_length"`
	ErrorMessage types.String `tfsdk:"error_message"`
}
type RegexValidator struct {
	Pattern      types.String `tfsdk:"pattern"`
	ErrorMessage types.String `tfsdk:"error_message"`
}
type EmailValidator struct {
	ErrorMessage types.String `tfsdk:"error_message"`
}

type AllowListValidator struct {
	Items        []types.String `tfsdk:"items"`
	ErrorMessage types.String   `tfsdk:"error_message"`
}

type DenyListValidator struct {
	Items        []types.String `tfsdk:"items"`
	ErrorMessage types.String   `tfsdk:"error_message"`
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
						"BOOLEAN", "TEXT", "NUMBER", "ISO_DATE", "SINGLE_SELECT_LIST", "MULTI_SELECT_LIST",
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
			"prechecked": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Check the checkbox(Boolean field) by default",
			},
			"placeholder": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The placeholder for TEXT or SINGLE_SELECT_LIST field",
			},
			"date_format": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The format for ISO_DATE field",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"mm/dd/yyyy",
						"mm.dd.yyyy",
						"dd/mm/yyyy",
						"dd.mm.yyyy",
						"yyyy/mm/dd",
						"yyyy.mm.dd",
						"yyyy/dd/mm",
						"yyyy.dd.mm",
					),
				},
			},
			"global": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether or not this field is a global field",
			},
			"multiline": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Piano ID custom field multiline setting for TEXT data type",
			},
			"pre_select_country_by_ip": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether or not select country by ip for country field. Default is false.",
			},
			"validators": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Piano ID custom field validators",
			},
			"length_validator": schema.ObjectAttribute{
				Optional:            true,
				MarkdownDescription: "Check if the input length fits between the min_length and max_length",
				AttributeTypes: map[string]attr.Type{
					"min_length":    types.Int32Type,
					"max_length":    types.Int32Type,
					"error_message": types.StringType,
				},
			},
			"regex_validator": schema.ObjectAttribute{
				Optional:            true,
				MarkdownDescription: "Check if the input matches the given regular expression",
				AttributeTypes: map[string]attr.Type{
					"pattern":       types.StringType,
					"error_message": types.StringType,
				},
			},
			"email_validator": schema.ObjectAttribute{
				Optional:            true,
				MarkdownDescription: "Check if the input conforms to valid email format",
				AttributeTypes: map[string]attr.Type{
					"error_message": types.StringType,
				},
			},
			"allow_list_validator": schema.ObjectAttribute{
				Optional:            true,
				MarkdownDescription: "Specify the allow list of possible inputs",
				AttributeTypes: map[string]attr.Type{
					"items": types.ListType{
						ElemType: types.StringType,
					},
					"error_message": types.StringType,
				},
			},
			"deny_list_validator": schema.ObjectAttribute{
				Optional:            true,
				MarkdownDescription: "Specify the deny list of possible inputs",
				AttributeTypes: map[string]attr.Type{
					"items": types.ListType{
						ElemType: types.StringType,
					},
					"error_message": types.StringType,
				},
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
	validators := validatorsFromState(state)
	favouriteOptions := favouriteOptionsFromState(state)
	request := piano_id.CustomFieldDefinition{
		FieldName:         state.FieldName.ValueString(),
		Title:             state.Title.ValueString(),
		Comment:           state.Comment.ValueStringPointer(),
		Editable:          state.Editable.ValueBool(),
		DataType:          piano_id.CustomFieldDefinitionDataType(state.DataType.ValueString()),
		Options:           options,
		FavouriteOptions:  &favouriteOptions,
		RequiredByDefault: state.RequiredByDefault.ValueBool(),
		Archived:          false,
		DefaultSortOrder:  state.DefaultSortOrder.ValueInt32Pointer(),
		Attribute: piano_id.CustomFieldAttribute{
			DefaultValue:         state.DefaultValue.ValueStringPointer(),
			Multiline:            state.Multiline.ValueBoolPointer(),
			Placeholder:          state.Placeholder.ValueStringPointer(),
			Global:               state.Global.ValueBoolPointer(),
			DateFormat:           (*piano_id.CustomFieldAttributeDateFormat)(state.DateFormat.ValueStringPointer()),
			PreSelectCountryByIp: state.PreSelectCountryById.ValueBoolPointer(),
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
	for _, validator := range data.Validators {
		if string(validator.Type) == "STR_LENGTH" && state.LengthValidator != nil {
			state.LengthValidator.MinLength = types.Int32PointerValue(validator.Params.MinLength)
			state.LengthValidator.MaxLength = types.Int32PointerValue(validator.Params.MaxLength)
			state.LengthValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else if string(validator.Type) == "REGEXP" && state.RegexValidator != nil {
			state.RegexValidator.Pattern = types.StringPointerValue(validator.Params.Regexp)
			state.RegexValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else if string(validator.Type) == "EMAIL" && state.EmailValidator != nil {
			state.EmailValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else if string(validator.Type) == "WHITELIST" && state.AllowListValidator != nil {
			items := []types.String{}
			if validator.Params.Whitelist != nil {
				for _, item := range *validator.Params.Whitelist {
					items = append(items, types.StringValue(item))
				}
			}
			state.AllowListValidator.Items = items
			state.AllowListValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else if string(validator.Type) == "BLACKLIST" && state.DenyListValidator != nil {
			items := []types.String{}
			if validator.Params.Blacklist != nil {
				for _, item := range *validator.Params.Blacklist {
					items = append(items, types.StringValue(item))
				}
			}

			state.DenyListValidator.Items = items
			state.DenyListValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else {
			// exaustiveness
		}
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
	validators := validatorsFromState(state)
	favouriteOptions := favouriteOptionsFromState(state)

	response, err := r.client.PublisherCustomFieldPost(ctx, []piano_id.CustomFieldDefinition{
		{
			FieldName:         state.FieldName.ValueString(),
			Title:             state.Title.ValueString(),
			Comment:           state.Comment.ValueStringPointer(),
			Editable:          state.Editable.ValueBool(),
			DataType:          piano_id.CustomFieldDefinitionDataType(state.DataType.ValueString()),
			Options:           options,
			FavouriteOptions:  &favouriteOptions,
			RequiredByDefault: state.RequiredByDefault.ValueBool(),
			Archived:          false,
			DefaultSortOrder:  state.DefaultSortOrder.ValueInt32Pointer(),
			Attribute: piano_id.CustomFieldAttribute{
				DefaultValue:         state.DefaultValue.ValueStringPointer(),
				Multiline:            state.Multiline.ValueBoolPointer(),
				Placeholder:          state.Placeholder.ValueStringPointer(),
				Global:               state.Global.ValueBoolPointer(),
				DateFormat:           (*piano_id.CustomFieldAttributeDateFormat)(state.DateFormat.ValueStringPointer()),
				PreSelectCountryByIp: state.PreSelectCountryById.ValueBoolPointer(),
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
	for _, validator := range data.Validators {
		if string(validator.Type) == "STR_LENGTH" && state.LengthValidator != nil {
			state.LengthValidator.MinLength = types.Int32PointerValue(validator.Params.MinLength)
			state.LengthValidator.MaxLength = types.Int32PointerValue(validator.Params.MaxLength)
			state.LengthValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else if string(validator.Type) == "REGEXP" && state.RegexValidator != nil {
			state.RegexValidator.Pattern = types.StringPointerValue(validator.Params.Regexp)
			state.RegexValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else if string(validator.Type) == "EMAIL" && state.EmailValidator != nil {
			state.EmailValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else if string(validator.Type) == "WHITELIST" && state.AllowListValidator != nil {
			items := []types.String{}
			if validator.Params.Whitelist != nil {
				for _, item := range *validator.Params.Whitelist {
					items = append(items, types.StringValue(item))
				}
			}

			state.AllowListValidator.Items = items
			state.AllowListValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else if string(validator.Type) == "BLACKLIST" && state.DenyListValidator != nil {
			items := []types.String{}
			if validator.Params.Blacklist != nil {
				for _, item := range *validator.Params.Blacklist {
					items = append(items, types.StringValue(item))
				}
			}

			state.DenyListValidator.Items = items
			state.DenyListValidator.ErrorMessage = types.StringPointerValue(validator.ReponseErrorMessage)
		} else {
			// exaustiveness
		}
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
	validators := validatorsFromState(state)
	favouriteOptions := favouriteOptionsFromState(state)

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
			FavouriteOptions:  &favouriteOptions,
			DefaultSortOrder:  state.DefaultSortOrder.ValueInt32Pointer(),
			Attribute: piano_id.CustomFieldAttribute{
				DefaultValue:         state.DefaultValue.ValueStringPointer(),
				Multiline:            state.Multiline.ValueBoolPointer(),
				Placeholder:          state.Placeholder.ValueStringPointer(),
				Global:               state.Global.ValueBoolPointer(),
				DateFormat:           (*piano_id.CustomFieldAttributeDateFormat)(state.DateFormat.ValueStringPointer()),
				PreSelectCountryByIp: state.PreSelectCountryById.ValueBoolPointer(),
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

func favouriteOptionsFromState(state CustomFieldResourceModel) []piano_id.CustomFieldDefinitionFavouriteOptions {
	options := []piano_id.CustomFieldDefinitionFavouriteOptions{}

	if state.Prechecked.ValueBool() {
		options = append(options, piano_id.Prechecked)
	}
	return options
}

func validatorsFromState(state CustomFieldResourceModel) []piano_id.Validator {
	validators := []piano_id.Validator{}
	if state.LengthValidator != nil {
		validators = append(validators, piano_id.Validator{
			Type: "STR_LENGTH",
			Params: piano_id.ValidatorParameter{
				MinLength: state.LengthValidator.MinLength.ValueInt32Pointer(),
				MaxLength: state.LengthValidator.MaxLength.ValueInt32Pointer(),
			},
			ErrorMessage: state.LengthValidator.ErrorMessage.ValueStringPointer(),
		})
	}
	if state.RegexValidator != nil {
		validators = append(validators, piano_id.Validator{
			Type: "REGEXP",
			Params: piano_id.ValidatorParameter{
				Regexp: state.RegexValidator.Pattern.ValueStringPointer(),
			},
			ErrorMessage: state.RegexValidator.ErrorMessage.ValueStringPointer(),
		})
	}
	if state.EmailValidator != nil {
		validators = append(validators, piano_id.Validator{
			Type:         "EMAIL",
			Params:       piano_id.ValidatorParameter{},
			ErrorMessage: state.EmailValidator.ErrorMessage.ValueStringPointer(),
		})
	}
	if state.AllowListValidator != nil {
		list := []string{}
		for _, item := range state.AllowListValidator.Items {
			list = append(list, item.ValueString())
		}
		validators = append(validators, piano_id.Validator{
			Type: "WHITELIST",
			Params: piano_id.ValidatorParameter{
				Whitelist: &list,
			},
			ErrorMessage: state.AllowListValidator.ErrorMessage.ValueStringPointer(),
		})
	}

	if state.DenyListValidator != nil {
		list := []string{}
		for _, item := range state.DenyListValidator.Items {
			list = append(list, item.ValueString())
		}
		validators = append(validators, piano_id.Validator{
			Type: "BLACKLIST",
			Params: piano_id.ValidatorParameter{
				Blacklist: &list,
			},
			ErrorMessage: state.DenyListValidator.ErrorMessage.ValueStringPointer(),
		})
	}
	return validators
}
