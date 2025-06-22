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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ContractSourceModel describes the data source data model.
type ContractResourceModel struct {
	// required
	Aid          types.String `tfsdk:"aid"`
	LicenseeId   types.String `tfsdk:"licensee_id"`
	ContractId   types.String `tfsdk:"contract_id"`
	ContractType types.String `tfsdk:"contract_type"`
	Rid          types.String `tfsdk:"rid"`
	Name         types.String `tfsdk:"name"`

	// Computed
	CreateDate       types.Int64 `tfsdk:"create_date"`
	ContractIsActive types.Bool  `tfsdk:"contract_is_active"`

	// Optional
	Description          types.String `tfsdk:"description"`
	LandingPageUrl       types.String `tfsdk:"landing_page_url"`
	SeatsNumber          types.Int32  `tfsdk:"seats_number"`
	IsHardSeatsLimitType types.Bool   `tfsdk:"is_hard_seats_limit_type"`
}

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &ContractResource{}
	_ resource.ResourceWithImportState = &ContractResource{}
)

func NewContractResource() resource.Resource {
	return &ContractResource{}
}

// ContractResource defines the resource implementation.
type ContractResource struct {
	client *piano_publisher.Client
}

func (*ContractResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contract"
}

func (*ContractResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Contract Resource. This resource is used to create, update, and delete a contract.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"contract_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The public ID of the contract",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the contract",
			},
			"licensee_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The public ID of the licensee",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description of the contract",
			},
			"contract_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of the contract. The value is one of the following: SPECIFIC_EMAIL_ADDRESSES_CONTRACT, EMAIL_DOMAIN_CONTRACT, IP_RANGE_CONTRACT",
				Validators: []validator.String{
					stringvalidator.OneOf("SPECIFIC_EMAIL_ADDRESSES_CONTRACT", "EMAIL_DOMAIN_CONTRACT", "IP_RANGE_CONTRACT"),
				},
			},
			"create_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The creation date of the contract",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"landing_page_url": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The relative URL of the contract. It will be appended to the licensing base URL to get the complete landing page URL",
			},
			"seats_number": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "The number of users who can access this contract",
			},
			"is_hard_seats_limit_type": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "The seats limit type (false: a notification is sent if the number of seats is exceeded, true: no user can access if the number of seats is exceeded)",
			},
			"contract_is_active": schema.BoolAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The seats limit type (false: a notification is sent if the number of seats is exceeded, true: no user can access if the number of seats is exceeded)",
			},
			"rid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The resource ID",
			},
		},
	}
}

func (r *ContractResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContractResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContractResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("creating contract %s in %s", plan.Name.ValueString(), plan.Aid.ValueString()))

	request := piano_publisher.PostPublisherLicensingContractCreateFormdataRequestBody{
		Aid:                  plan.Aid.ValueString(),
		LicenseeId:           plan.LicenseeId.ValueString(),
		ContractType:         piano_publisher.PostPublisherLicensingContractCreateRequestContractType(plan.ContractType.ValueString()),
		ContractName:         plan.Name.ValueString(),
		SeatsNumber:          plan.SeatsNumber.ValueInt32(),
		IsHardSeatsLimitType: plan.IsHardSeatsLimitType.ValueBool(),
		Rid:                  plan.Rid.ValueString(),
		LandingPageUrl:       plan.LandingPageUrl.ValueStringPointer(),
	}

	response, err := r.client.PostPublisherLicensingContractCreateWithFormdataBody(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.ContractResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}
	// Computed
	plan.ContractId = types.StringValue(result.Contract.ContractId)
	plan.CreateDate = types.Int64Value(int64(result.Contract.CreateDate))
	plan.ContractIsActive = types.BoolValue(result.Contract.ContractIsActive)
	// Updated
	plan.ContractType = types.StringValue(string(result.Contract.ContractType))
	plan.LicenseeId = types.StringValue(result.Contract.LicenseeId)
	plan.Rid = types.StringValue(result.Contract.Rid)
	plan.Name = types.StringValue(result.Contract.Name)
	plan.Description = types.StringPointerValue(result.Contract.Description)
	plan.IsHardSeatsLimitType = types.BoolValue(result.Contract.IsHardSeatsLimitType)
	plan.SeatsNumber = types.Int32Value(result.Contract.SeatsNumber)
	plan.LandingPageUrl = types.StringValue(result.Contract.LandingPageUrl)
	tflog.Info(ctx, fmt.Sprintf("complete creating contract %s(id: %s)", plan.Name, plan.ContractId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContractResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContractResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.GetPublisherLicensingContractGet(ctx, &piano_publisher.GetPublisherLicensingContractGetParams{
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

	result := piano_publisher.ContractResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	// Populate state with the response data
	state.Rid = types.StringValue(result.Contract.Rid)
	state.Name = types.StringValue(result.Contract.Name)
	state.Description = types.StringPointerValue(result.Contract.Description)
	state.CreateDate = types.Int64Value(int64(result.Contract.CreateDate))
	state.ContractIsActive = types.BoolValue(result.Contract.ContractIsActive)
	state.ContractType = types.StringValue(string(result.Contract.ContractType))
	state.LandingPageUrl = types.StringValue(result.Contract.LandingPageUrl)
	state.LicenseeId = types.StringValue(result.Contract.LicenseeId)
	state.SeatsNumber = types.Int32Value(result.Contract.SeatsNumber)
	state.IsHardSeatsLimitType = types.BoolValue(result.Contract.IsHardSeatsLimitType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ContractResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state ContractResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("updating contract %s(id:%s) in %s", state.Name.ValueString(), state.ContractId.ValueString(), state.Aid.ValueString()))

	request := piano_publisher.PostPublisherLicensingContractUpdateFormdataRequestBody{
		Aid:                  state.Aid.ValueString(),
		Rid:                  state.Rid.ValueString(),
		ContractId:           state.ContractId.ValueString(),
		LicenseeId:           state.LicenseeId.ValueString(),
		SeatsNumber:          state.SeatsNumber.ValueInt32(),
		IsHardSeatsLimitType: state.IsHardSeatsLimitType.ValueBool(),
		ContractName:         state.Name.ValueString(),
		LandingPageUrl:       state.LandingPageUrl.ValueString(),
		ContractType:         piano_publisher.EMAILDOMAINCONTRACT,
	}
	if state.LandingPageUrl.ValueString() != "" {
		request.LandingPageUrl = state.LandingPageUrl.ValueString()
	}
	if state.Description.ValueString() != "" {
		request.ContractDescription = state.Description.ValueStringPointer()
	}
	tflog.Info(ctx, fmt.Sprintf("request: %+v", state))

	response, err := r.client.PostPublisherLicensingContractUpdateWithFormdataBody(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update contract, got error: %s", err))
		tflog.Error(ctx, fmt.Sprintf("Unable to update contract: %e", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}
	result := piano_publisher.ContractResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}
	// Computed
	state.ContractIsActive = types.BoolValue(result.Contract.ContractIsActive)
	state.CreateDate = types.Int64Value(int64(result.Contract.CreateDate))
	// Updatable
	state.LicenseeId = types.StringValue(result.Contract.LicenseeId)
	state.ContractType = types.StringValue(string(result.Contract.ContractType))
	state.Rid = types.StringValue(result.Contract.Rid)
	state.Name = types.StringValue(result.Contract.Name)
	state.Description = types.StringPointerValue(result.Contract.Description)
	state.SeatsNumber = types.Int32Value(result.Contract.SeatsNumber)
	state.IsHardSeatsLimitType = types.BoolValue(result.Contract.IsHardSeatsLimitType)
	state.SeatsNumber = types.Int32Value(result.Contract.SeatsNumber)
	state.LandingPageUrl = types.StringValue(result.Contract.LandingPageUrl)
	tflog.Info(ctx, fmt.Sprintf("complete updating contract %s(id: %s)", state.Name, state.ContractId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ContractResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContractResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("deleting contract %s:%s in $%s", state.Name.ValueString(), state.ContractId.ValueString(), state.Aid.ValueString()))
	response, err := r.client.PostPublisherLicensingContractArchiveWithFormdataBody(ctx, piano_publisher.PostPublisherLicensingContractArchiveFormdataRequestBody{
		Aid:        state.Aid.ValueString(),
		ContractId: state.ContractId.ValueString(),
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

func (r *ContractResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceId, err := ContractResourceIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Contract resource id", fmt.Sprintf("Unable to parse contract resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), resourceId.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("contract_id"), resourceId.ContractId)...)
}

// ContractResourceId represents a piano.io contract resource identifier in "{aid}/{contract_id}" format.
type ContractResourceId struct {
	Aid        string
	ContractId string
}

func ContractResourceIdFromString(input string) (*ContractResourceId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, errors.New("contract resource id must be in {aid}/{contract_id} format")
	}
	return &ContractResourceId{Aid: parts[0], ContractId: parts[1]}, nil
}
