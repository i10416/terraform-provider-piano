// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TermChangeOptionV2ResourceModel struct {
	TermChangeOptionId types.String `tfsdk:"term_change_option_id"`
	Aid                types.String `tfsdk:"aid"`            // The application ID
	FromTermId         types.String `tfsdk:"from_term_id"`   // The term ID to change from
	ToTermId           types.String `tfsdk:"to_term_id"`     // The term ID to change to
	BillingTiming      types.String `tfsdk:"billing_timing"` // The billing timing
	ImmediateAccess    types.Bool   `tfsdk:"immediate_access"`
	ProrateAccess      types.Bool   `tfsdk:"prorate_access"`
	Description        types.String `tfsdk:"description"` // The description
}

var (
	_ resource.Resource                = &TermChangeOptionResource{}
	_ resource.ResourceWithImportState = &TermChangeOptionResource{}
)

func NewTermChangeOptionResource() resource.Resource {
	return &TermChangeOptionResource{}
}

// TermDataSource defines the data source implementation.
type TermChangeOptionResource struct {
	client *piano_publisher.Client
}

func (r *TermChangeOptionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_term_change_option"
}

func (r *TermChangeOptionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = &client.publisherClient
}

func (*TermChangeOptionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Payment Term Change Option resource.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"term_change_option_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The term change option ID",
			},
			"from_term_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The term ID to change from",
			},
			"to_term_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The term ID to change to",
			},
			"billing_timing": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("0", "1", "2", "3"),
				},
				MarkdownDescription: "The billing timing",
			},
			"immediate_access": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Immediate access flag",
			},
			"prorate_access": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Prorate access flag",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description",
			},
		},
	}
}

func (r *TermChangeOptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TermChangeOptionV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.PostPublisherTermChangeOptionCreateWithFormdataBody(ctx, piano_publisher.PostPublisherTermChangeOptionCreateFormdataRequestBody{
		Aid:             plan.Aid.ValueString(),
		FromTermId:      plan.FromTermId.ValueString(),
		ToTermId:        plan.ToTermId.ValueString(),
		BillingTiming:   piano_publisher.PostPublisherTermChangeOptionCreateRequestBillingTiming(plan.BillingTiming.ValueString()),
		ImmediateAccess: plan.ImmediateAccess.ValueBool(),
		ProrateAccess:   plan.ProrateAccess.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create resource, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	tflog.Info(ctx, "created Payment Term Change Option")
	result := piano_publisher.TermChangeOptionResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}
	option := TermChangeOptionV2ResourceModelFrom(result.TermChangeOption)
	tflog.Info(ctx, fmt.Sprintf("created Term Change Option %s from %s to %s", option.TermChangeOptionId.ValueString(), option.FromTermId.ValueString(), option.ToTermId.ValueString()))
	plan.TermChangeOptionId = option.TermChangeOptionId
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}
func (r *TermChangeOptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Warn(ctx, "updating a resource is not supported for Term Change Option resource")
}

func (r *TermChangeOptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TermChangeOptionV2ResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.GetPublisherTermGet(ctx, &piano_publisher.GetPublisherTermGetParams{
		TermId: state.FromTermId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch term, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.TermResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	data := result.Term
	state.Aid = types.StringValue(data.Aid)
	var ExpectedTermChangeOption *piano_publisher.TermChangeOption
	for _, termChangeOption := range data.ChangeOptions {
		if termChangeOption.TermChangeOptionId == state.TermChangeOptionId.ValueString() {
			ExpectedTermChangeOption = &termChangeOption
			break
		}
	}
	if ExpectedTermChangeOption == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Term Change Option %s not found in Term %s", state.TermChangeOptionId.ValueString(), state.FromTermId.ValueString()))
		return
	}
	state.TermChangeOptionId = types.StringValue(ExpectedTermChangeOption.TermChangeOptionId)
	state.Aid = types.StringValue(data.Aid)
	state.FromTermId = types.StringValue(ExpectedTermChangeOption.FromTermId)
	state.ToTermId = types.StringValue(ExpectedTermChangeOption.ToTermId)
	state.BillingTiming = types.StringValue(string(ExpectedTermChangeOption.BillingTiming))
	state.ImmediateAccess = types.BoolValue(ExpectedTermChangeOption.ImmediateAccess)
	state.ProrateAccess = types.BoolValue(ExpectedTermChangeOption.ProrateAccess)
	if ExpectedTermChangeOption.Description != "" {
		state.Description = types.StringValue(ExpectedTermChangeOption.Description)
	}

	tflog.Trace(ctx, "read a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TermChangeOptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Warn(ctx, "deleting a resource is not supported for Term Change Option resource")
}

func (r *TermChangeOptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := TermChangeOptionV2ResourceIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Term Change Option resource id", fmt.Sprintf("Unable to parse contract resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), id.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("from_term_id"), id.TermId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("term_change_option_id"), id.TermChangeOptionId)...)
}
func TermChangeOptionV2ResourceIdFromString(input string) (*TermChangeOptionV2ResourceId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 3 {
		return nil, errors.New("term resource id must be in {aid}/{term_id} format")
	}
	return &TermChangeOptionV2ResourceId{Aid: parts[0], TermId: parts[1], TermChangeOptionId: parts[2]}, nil
}

type TermChangeOptionV2ResourceId struct {
	Aid                string
	TermId             string
	TermChangeOptionId string
}

func TermChangeOptionV2ResourceModelFrom(data piano_publisher.TermChangeOption) TermChangeOptionV2ResourceModel {
	ret := TermChangeOptionV2ResourceModel{}
	ret.TermChangeOptionId = types.StringValue(data.TermChangeOptionId)
	ret.BillingTiming = types.StringValue(string(data.BillingTiming))
	ret.Description = types.StringValue(data.Description)
	ret.FromTermId = types.StringValue(data.FromTermId)
	ret.ImmediateAccess = types.BoolValue(data.ImmediateAccess)
	ret.ToTermId = types.StringValue(data.ToTermId)
	ret.ProrateAccess = types.BoolValue(data.ProrateAccess)
	return ret
}
