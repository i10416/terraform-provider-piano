// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &OfferTermBindingResource{}
	_ resource.ResourceWithImportState = &OfferTermBindingResource{}
)

type OfferTermBindingResource struct {
	client *piano_publisher.Client
}

func NewOfferTermBindingResource() resource.Resource {
	return &OfferTermBindingResource{}
}
func (r *OfferTermBindingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*piano_publisher.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *piano_publisher.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *OfferTermBindingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_offer_term_binding"
}

type OfferTermBindingResourceModel struct {
	Aid     types.String `tfsdk:"aid"`      // The application ID
	OfferId types.String `tfsdk:"offer_id"` // The offer ID
	TermId  types.String `tfsdk:"term_id"`
}

func (*OfferTermBindingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OfferTermBinding resource associates a term with an offer",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"offer_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The offer ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"term_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The term id in the offer",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *OfferTermBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OfferTermBindingResourceModel
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
	present := slices.ContainsFunc(data, func(term piano_publisher.Term) bool {
		return term.TermId == state.TermId.ValueString()
	})
	if present {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.Diagnostics.AddError("Resource not found", fmt.Sprintf("The offer(%s) does not contain the term(%s)", state.OfferId, state.TermId))
		return
	}
}
func (r *OfferTermBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state OfferTermBindingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.PostPublisherOfferTermAddWithFormdataBody(ctx, piano_publisher.PostPublisherOfferTermAddFormdataRequestBody{
		Aid:     state.Aid.ValueString(),
		OfferId: state.OfferId.ValueString(),
		TermId:  state.TermId.ValueString(),
	})
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
func (r *OfferTermBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}
func (r *OfferTermBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OfferTermBindingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.PostPublisherOfferTermRemoveWithFormdataBody(ctx, piano_publisher.PostPublisherOfferTermRemoveFormdataRequestBody{
		Aid:     state.Aid.ValueString(),
		OfferId: state.OfferId.ValueString(),
		TermId:  state.TermId.ValueString(),
	})
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

func (r *OfferTermBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := OfferTermBindingIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid offer term resource id", fmt.Sprintf("Unable to parse offer resource id, got error: %e", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), id.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("offer_id"), id.OfferId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("term_id"), id.TermId)...)
}

type OfferTermBindingResourceId struct {
	Aid     string
	OfferId string
	TermId  string
}

func OfferTermBindingIdFromString(input string) (*OfferTermBindingResourceId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 3 {
		return nil, errors.New("offer term resource id must be in {aid}/{offer_id}/{term_id} format")
	}
	data := OfferTermBindingResourceId{
		Aid:     parts[0],
		OfferId: parts[1],
		TermId:  parts[2],
	}
	return &data, nil
}
