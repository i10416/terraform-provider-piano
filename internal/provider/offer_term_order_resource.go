// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OfferTermOrderResource{}
	_ resource.ResourceWithImportState = &OfferTermOrderResource{}
)

type OfferTermOrderResource struct {
	client *piano_publisher.Client
}

func NewOfferTermOrderResource() resource.Resource {
	return &OfferTermOrderResource{}
}
func (r *OfferTermOrderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(PianoProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *piano_publisher.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = &client.publisherClient
}

func (r *OfferTermOrderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_offer_term_order"
}

type OfferTermOrderResourceModel struct {
	Aid     types.String `tfsdk:"aid"`      // The application ID
	OfferId types.String `tfsdk:"offer_id"` // The offer ID
	TermIds []string     `tfsdk:"term_ids"`
}

func (*OfferTermOrderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource defines the order of terms in an offer.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"offer_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The offer ID",
			},
			"term_ids": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "The term ids in the offer",
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *OfferTermOrderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OfferTermOrderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.GetPublisherOfferTermList(ctx, &piano_publisher.GetPublisherOfferTermListParams{
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

	result := piano_publisher.TermArrayResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	data := result.Terms
	termIds := []string{}
	for _, term := range data {
		termIds = append(termIds, term.TermId)
	}

	state.TermIds = termIds

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *OfferTermOrderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state OfferTermOrderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data := url.Values{}
	data.Set("aid", state.Aid.ValueString())
	data.Set("offer_id", state.OfferId.ValueString())
	for _, id := range state.TermIds {
		data.Add("term_id", id)
	}
	body := strings.NewReader(data.Encode())
	response, err := r.client.PostPublisherOfferTermReorderWithBody(ctx, "application/x-www-form-urlencoded", body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create offer, got error: %s", err))
		return
	}
	_, err = syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *OfferTermOrderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state OfferTermOrderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data := url.Values{}
	data.Set("aid", state.Aid.ValueString())
	data.Set("offer_id", state.OfferId.ValueString())
	for _, id := range state.TermIds {
		data.Add("term_id", id)
	}
	body := strings.NewReader(data.Encode())
	response, err := r.client.PostPublisherOfferTermReorderWithBody(ctx, "application/x-www-form-urlencoded", body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update offer, got error: %s", err))
		return
	}
	_, err = syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *OfferTermOrderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OfferTermOrderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *OfferTermOrderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := OfferIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid offer resource id", fmt.Sprintf("Unable to parse offer resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), id.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("offer_id"), id.OfferId)...)
}
