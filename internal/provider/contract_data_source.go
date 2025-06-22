// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ContractDataSource{}
	_ datasource.DataSourceWithConfigure = &ContractDataSource{}
)

func NewContractDataSource() datasource.DataSource {
	return &ContractDataSource{}
}

// ContractDataSource defines the data source implementation.
type ContractDataSource struct {
	client *piano_publisher.Client
}

// SchedulePeriodModel describes the schedule period data model.
type SchedulePeriodModel struct {
	PeriodId  types.String `tfsdk:"period_id"`
	Name      types.String `tfsdk:"name"`
	SellDate  types.Int64  `tfsdk:"sell_date"`
	BeginDate types.Int64  `tfsdk:"begin_date"`
	EndDate   types.Int64  `tfsdk:"end_date"`
	Status    types.String `tfsdk:"status"`
}

// ContractDataSourceModel describes the data source data model.
type ContractDataSourceModel struct {
	// required
	Aid        types.String `tfsdk:"aid"`
	ContractId types.String `tfsdk:"contract_id"`

	// computed
	Name                     types.String          `tfsdk:"name"`
	Description              types.String          `tfsdk:"description"`
	CreateDate               types.Int64           `tfsdk:"create_date"`
	LandingPageUrl           types.String          `tfsdk:"landing_page_url"`
	LicenseeId               types.String          `tfsdk:"licensee_id"`
	SeatsNumber              types.Int32           `tfsdk:"seats_number"`
	IsHardSeatsLimitType     types.Bool            `tfsdk:"is_hard_seats_limit_type"`
	Rid                      types.String          `tfsdk:"rid"`
	ContractIsActive         types.Bool            `tfsdk:"contract_is_active"`
	ContractType             types.String          `tfsdk:"contract_type"`
	ContractConversionsCount types.Int32           `tfsdk:"contract_conversions_count"`
	ContractPeriods          []SchedulePeriodModel `tfsdk:"contract_periods"`
}

func (d *ContractDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contract"
}

func (d *ContractDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Contract data source. This data source is used to get the contract details.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"contract_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The public ID of the contract",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the contract",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The description of the contract",
			},
			"create_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The creation date of the contract",
			},
			"landing_page_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The relative URL of the contract. It will be appended to the licensing base URL to get the complete landing page URL",
			},
			"licensee_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The public ID of the licensee",
			},
			"seats_number": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The number of users who can access this contract",
			},
			"is_hard_seats_limit_type": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "The seats limit type (false: a notification is sent if the number of seats is exceeded, true: no user can access if the number of seats is exceeded)",
			},
			"rid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resource ID",
			},
			"contract_is_active": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "The contract is active",
			},
			"contract_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of the contract",
			},
			"contract_conversions_count": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The number of conversions for the contract",
			},
			"contract_periods": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The periods of the contract",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"period_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the period",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the period",
						},
						"sell_date": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The sell date of the period",
						},
						"begin_date": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The begin date of the period",
						},
						"end_date": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The end date of the period",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The status of the period",
						},
					},
				},
			},
		},
	}
}

func (d *ContractDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(PianoProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *piano_publisher.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = &client.publisherClient
}

func (d *ContractDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ContractDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := d.client.GetPublisherLicensingContractGet(ctx, &piano_publisher.GetPublisherLicensingContractGetParams{
		Aid:        state.Aid.ValueString(),
		ContractId: state.ContractId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch licensee, got error: %s", err))
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

	state.LandingPageUrl = types.StringValue(result.Contract.LandingPageUrl)
	state.LicenseeId = types.StringValue(result.Contract.LicenseeId)

	state.SeatsNumber = types.Int32Value(result.Contract.SeatsNumber)
	state.IsHardSeatsLimitType = types.BoolValue(result.Contract.IsHardSeatsLimitType)

	state.ContractIsActive = types.BoolValue(result.Contract.ContractIsActive)
	state.ContractType = types.StringValue(string(result.Contract.ContractType))
	state.ContractConversionsCount = types.Int32Value(result.Contract.ContractConversionsCount)

	state.ContractPeriods = SchedulePeriodDataSourceFromData(result.Contract.ContractPeriods)

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func SchedulePeriodDataSourceFromData(items []piano_publisher.SchedulePeriod) []SchedulePeriodModel {
	periods := []SchedulePeriodModel{}
	for _, item := range items {
		period := SchedulePeriodModel{}
		period.PeriodId = types.StringValue(item.PeriodId)
		period.Name = types.StringValue(item.Name)
		period.SellDate = types.Int64Value(int64(item.SellDate))
		period.BeginDate = types.Int64Value(int64(item.BeginDate))
		period.EndDate = types.Int64Value(int64(item.EndDate))
		period.Status = types.StringValue(string(item.Status))
		periods = append(periods, period)
	}
	return periods
}
