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

// ContractDomainSourceModel describes the data source data model.
type ContractDomainResourceModel struct {
	// required
	Aid                 types.String `tfsdk:"aid"`
	ContractDomainId    types.String `tfsdk:"contract_domain_id"`
	ContractId          types.String `tfsdk:"contract_id"`
	ContractDomainValue types.String `tfsdk:"contract_domain_value"`
}

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &ContractDomainResource{}
	_ resource.ResourceWithImportState = &ContractDomainResource{}
)

func NewContractDomainResource() resource.Resource {
	return &ContractDomainResource{}
}

// ContractDomainResource defines the resource implementation.
type ContractDomainResource struct {
	client *piano_publisher.Client
}

func (*ContractDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contract_domain"
}

func (*ContractDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ContractDomain Resource. This resource is used to create, update, and delete a contract domain.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"contract_domain_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The id of the contract domain",
			},
			"contract_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The public ID of the contract",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"contract_domain_value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The value of the contract domain",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ContractDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContractDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContractDomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("creating contract %s in %s", plan.ContractDomainValue.ValueString(), plan.Aid.ValueString()))

	request := piano_publisher.PostPublisherLicensingContractDomainCreateFormdataRequestBody{
		Aid:                 plan.Aid.ValueString(),
		ContractId:          plan.ContractId.ValueString(),
		ContractDomainValue: plan.ContractDomainValue.ValueString(),
	}

	response, err := r.client.PostPublisherLicensingContractDomainCreateWithFormdataBody(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.ContractDomainResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}
	// Computed
	plan.ContractDomainId = types.StringValue(result.ContractDomain.ContractDomainId)
	tflog.Info(ctx, fmt.Sprintf("complete creating contract %s(id: %s)", plan.ContractDomainValue, plan.ContractDomainId))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContractDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContractDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.GetPublisherLicensingContractDomainList(ctx, &piano_publisher.GetPublisherLicensingContractDomainListParams{
		Aid:        state.Aid.ValueString(),
		ContractId: state.ContractId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch contract, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.ContractDomainArrayResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}
	var domain *piano_publisher.ContractDomain
	for _, item := range result.ContractDomainList {
		if item.ContractDomainId == state.ContractDomainId.ValueString() {
			domain = &item
		}
	}
	if domain == nil {
		resp.Diagnostics.AddError("Not Found Error", fmt.Sprintf("Unable to find piano contract domain: %s with id: %s", state.ContractDomainValue, state.ContractDomainId))
		return
	}

	state.ContractDomainId = types.StringValue(domain.ContractDomainId)
	state.ContractDomainValue = types.StringValue(domain.ContractDomainValue)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ContractDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state ContractDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ContractDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContractDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.PostPublisherLicensingContractDomainRemoveWithFormdataBody(ctx, piano_publisher.PostPublisherLicensingContractDomainRemoveFormdataRequestBody{
		Aid:              state.Aid.ValueString(),
		ContractId:       state.ContractId.ValueString(),
		ContractDomainId: state.ContractDomainId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete contract, got error: %s", err))
		return
	}
	_, err = syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}
}

func (r *ContractDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceId, err := ContractDomainResourceIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ContractDomain resource id", fmt.Sprintf("Unable to parse contract resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), resourceId.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("contract_id"), resourceId.ContractId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("contract_domain_id"), resourceId.ContractDomainId)...)
}

// ContractDomainResourceId represents a piano.io contract resource identifier in "{aid}/{contract_domain_id}" format.
type ContractDomainResourceId struct {
	Aid              string
	ContractId       string
	ContractDomainId string
}

func ContractDomainResourceIdFromString(input string) (*ContractDomainResourceId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 3 {
		return nil, errors.New("contract domain id must be in {aid}/{contract_id}/{contract_domain_id} format")
	}
	return &ContractDomainResourceId{Aid: parts[0], ContractId: parts[1], ContractDomainId: parts[2]}, nil
}
