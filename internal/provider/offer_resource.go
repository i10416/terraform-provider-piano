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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &OfferResource{}
	_ resource.ResourceWithImportState = &OfferResource{}
)

type OfferResource struct {
	client *piano_publisher.Client
}

func NewOfferResource() resource.Resource {
	return &OfferResource{}
}
func (r *OfferResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OfferResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_offer"
}

type OfferResourceModel struct {
	Name    types.String `tfsdk:"name"`     // The offer name
	Aid     types.String `tfsdk:"aid"`      // The application ID
	OfferId types.String `tfsdk:"offer_id"` // The offer ID
}

func (*OfferResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"offer_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The offer ID",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The offer name",
			},
		},
	}
}

func (r *OfferResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OfferResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.GetPublisherOfferGet(ctx, &piano_publisher.GetPublisherOfferGetParams{
		Aid:     state.Aid.ValueString(),
		OfferId: state.OfferId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch promotion, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.OfferModelResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	data := result.Offer
	state.OfferId = types.StringValue(data.OfferId)
	state.Aid = types.StringValue(data.Aid)

	state.Name = types.StringValue(data.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *OfferResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state OfferResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.PostPublisherOfferCreateWithFormdataBody(ctx, piano_publisher.PostPublisherOfferCreateFormdataRequestBody{
		Aid:  state.Aid.ValueString(),
		Name: state.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create offer, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.OfferModelResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %e", err))
		return
	}

	data := result.Offer

	state.OfferId = types.StringValue(data.OfferId)
	state.Aid = types.StringValue(data.Aid)

	state.Name = types.StringValue(data.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *OfferResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state OfferResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.PostPublisherOfferUpdateWithFormdataBody(ctx, piano_publisher.PostPublisherOfferUpdateFormdataRequestBody{
		Aid:  state.Aid.ValueString(),
		Name: state.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update offer, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.OfferModelResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	data := result.Offer
	state.OfferId = types.StringValue(data.OfferId)
	state.Aid = types.StringValue(data.Aid)
	state.Name = types.StringValue(data.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *OfferResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OfferResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, fmt.Sprintf("deleting promotion %s:%s in $%s", state.Name.ValueString(), state.OfferId.ValueString(), state.Aid.ValueString()))
	response, err := r.client.PostPublisherOfferDeleteWithFormdataBody(ctx, piano_publisher.PostPublisherOfferDeleteFormdataRequestBody{
		Aid:     state.Aid.ValueString(),
		OfferId: state.OfferId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete licensee, got error: %s", err))
		return
	}
	_, err = syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}
}
func (r *OfferResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := OfferIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid offer resource id", fmt.Sprintf("Unable to parse offer resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), id.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("offer_id"), id.OfferId)...)
}

// OfferId represents a piano.io promotion resource identifier in "{aid}/{offer_id}" format.
type OfferResourceId struct {
	Aid     string
	OfferId string
}

func OfferIdFromString(input string) (*OfferResourceId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, errors.New("offer resource id must be in {aid}/{offer_id} format")
	}
	data := OfferResourceId{
		Aid:     parts[0],
		OfferId: parts[1],
	}
	return &data, nil
}
