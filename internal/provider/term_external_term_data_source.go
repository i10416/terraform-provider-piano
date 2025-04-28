// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ExternalTermDataSourceModel struct {
	Aid                      types.String                      `tfsdk:"aid"`                          // The application ID
	CollectAddress           types.Bool                        `tfsdk:"collect_address"`              // Whether to collect an address for this term
	CreateDate               types.Int64                       `tfsdk:"create_date"`                  // The creation date
	Description              types.String                      `tfsdk:"description"`                  // The description of the term
	EvtFixedTimeAccessPeriod types.Int32                       `tfsdk:"evt_fixed_time_access_period"` // The period to grant access for (in days)
	EvtGooglePlayProductId   types.String                      `tfsdk:"evt_google_play_product_id"`   // Google Play's product ID
	EvtGracePeriod           types.Int32                       `tfsdk:"evt_grace_period"`             // The External API grace period
	EvtItunesBundleId        types.String                      `tfsdk:"evt_itunes_bundle_id"`         // iTunes's bundle ID
	EvtItunesProductId       types.String                      `tfsdk:"evt_itunes_product_id"`        // iTunes's product ID
	EvtVerificationPeriod    types.Int32                       `tfsdk:"evt_verification_period"`      // The <a href = "https://docs.piano.io/external-service-term/#externaltermverification">periodicity</a> (in seconds) of checking the EVT subscription with the external service
	ExternalApiFormFields    []ExternalAPIFieldDataSourceModel `tfsdk:"external_api_form_fields"`
	ExternalApiId            types.String                      `tfsdk:"external_api_id"`     // The ID of the external API configuration
	ExternalApiName          types.String                      `tfsdk:"external_api_name"`   // The name of the external API configuration
	ExternalApiSource        types.Int32                       `tfsdk:"external_api_source"` // The source of the external API configuration
	Name                     types.String                      `tfsdk:"name"`                // The term name
	Resource                 *ResourceDataSourceModel          `tfsdk:"resource"`
	SharedAccountCount       types.Int32                       `tfsdk:"shared_account_count"`  // The count of allowed shared-subscription accounts
	SharedRedemptionUrl      types.String                      `tfsdk:"shared_redemption_url"` // The shared subscription redemption URL
	TermId                   types.String                      `tfsdk:"term_id"`               // The term ID
	Type                     types.String                      `tfsdk:"type"`                  // The term type
	TypeName                 types.String                      `tfsdk:"type_name"`             // The term type name
	UpdateDate               types.Int64                       `tfsdk:"update_date"`           // The update date
}

func (r *ExternalTermDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_term"
}
func (*ExternalTermDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ExternalTerm datasource. External term is a term that is created by the external API.",
		Attributes: map[string]schema.Attribute{
			"term_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The term ID",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The description of the term",
			},
			"collect_address": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to collect an address for this term",
			},
			"evt_itunes_bundle_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "iTunes's bundle ID",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The term name",
			},
			"evt_itunes_product_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "iTunes's product ID",
			},
			"shared_account_count": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The count of allowed shared-subscription accounts",
			},
			"external_api_form_fields": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mandatory": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the field is required",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The field description, some information about what information should be entered",
						},
						"hidden": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the field will be submitted hiddenly from the user, default value is required",
						},
						"order": schema.Int32Attribute{
							Computed:            true,
							MarkdownDescription: "Field order in the list",
						},
						"default_value": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Default value for the field. It will be pre-entered on the form",
						},
						"field_title": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The title of the field to be displayed to the user",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Field type",
							Validators: []validator.String{
								stringvalidator.OneOf("INPUT", "COUNTRY_SELECTOR", "STATE_AUTOCOMPLETE"),
							},
						},
						"field_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the field to be used to submit to the external system",
						},
						"editable": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the object is editable",
						},
					},
				},
			},
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"external_api_source": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The source of the external API configuration",
			},
			"external_api_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the external API configuration",
			},
			"update_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The update date",
			},
			"create_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The creation date",
			},
			"evt_verification_period": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The <a href = \"https://docs.piano.io/external-service-term/#externaltermverification\">periodicity</a> (in seconds) of checking the EVT subscription with the external service",
			},
			"evt_google_play_product_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Google Play's product ID",
			},
			"resource": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"bundle_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The resource bundle type",
						Validators: []validator.String{
							stringvalidator.OneOf("undefined", "fixed", "tagged", "fixed_v2"),
						},
					},
					"image_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The URL of the resource image",
					},
					"purchase_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The URL of the purchase page",
					},
					"aid": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The application ID",
					},
					"description": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The resource description",
					},
					"update_date": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The update date",
					},
					"bundle_type_label": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The bundle type label",
						Validators: []validator.String{
							stringvalidator.OneOf("Undefined", "Fixed", "Tagged", "Fixed 2.0"),
						},
					},
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The type of the resource (0: Standard, 4: Bundle)",
						Validators: []validator.String{
							stringvalidator.OneOf("standard", "bundle", "print"),
						},
					},
					"deleted": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether the object is deleted",
					},
					"rid": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The resource ID",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The name",
					},
					"create_date": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The creation date",
					},
					"is_fbia_resource": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Enable the resource for Facebook Subscriptions in Instant Articles",
					},
					"type_label": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The resource type label (\"Standard\" or \"Bundle\")",
						Validators: []validator.String{
							stringvalidator.OneOf("Standard", "Bundle", "Print"),
						},
					},
					"external_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The external ID; defined by the client",
					},
					"publish_date": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The publish date",
					},
					"resource_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The URL of the resource",
					},
					"disabled": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether the object is disabled",
					},
				},
			},
			"evt_fixed_time_access_period": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The period to grant access for (in days)",
			},
			"evt_grace_period": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The External API grace period",
			},
			"type_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The term type name",
				Validators: []validator.String{
					stringvalidator.OneOf("Payment", "Ad View", "Registration", "Newsletter", "External", "Custom", "Access Granted", "Gift", "Specific Email Addresses Contract", "Email Domain Contract", "IP Range Contract", "Dynamic", "Linked"),
				},
			},
			"shared_redemption_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The shared subscription redemption URL",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The term type",
				Validators: []validator.String{
					stringvalidator.OneOf("payment", "adview", "registration", "newsletter", "external", "custom", "grant_access", "gift", "specific_email_addresses_contract", "email_domain_contract", "ip_range_contract", "dynamic", "linked"),
				},
			},
			"external_api_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the external API configuration",
			},
		},
	}
}

var (
	_ datasource.DataSource              = &ExternalTermDataSource{}
	_ datasource.DataSourceWithConfigure = &ExternalTermDataSource{}
)

func NewExternalTermDataSource() datasource.DataSource {
	return &ExternalTermDataSource{}
}

// TermDataSource defines the data source implementation.
type ExternalTermDataSource struct {
	client *piano_publisher.Client
}

func (r *ExternalTermDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *ExternalTermDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ExternalTermDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.GetPublisherTermGet(ctx, &piano_publisher.GetPublisherTermGetParams{
		TermId: state.TermId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch term, got error: %s", err))
		return
	}
	anyResponse, err := syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.ExternalTermResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	data := result.Term
	state.ExternalApiId = types.StringValue(data.ExternalApiId)
	state.Type = types.StringValue(string(data.Type))
	state.SharedRedemptionUrl = types.StringPointerValue(data.SharedRedemptionUrl)
	state.TypeName = types.StringValue(string(data.TypeName))
	state.EvtGracePeriod = types.Int32Value(data.EvtGracePeriod)
	state.EvtFixedTimeAccessPeriod = types.Int32PointerValue(data.EvtFixedTimeAccessPeriod)
	Resource := ResourceDataSourceModelFrom(data.Resource)
	state.Resource = &Resource
	state.EvtGooglePlayProductId = types.StringPointerValue(data.EvtGooglePlayProductId)
	state.EvtVerificationPeriod = types.Int32PointerValue(data.EvtVerificationPeriod)
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	state.ExternalApiName = types.StringValue(data.ExternalApiName)
	state.ExternalApiSource = types.Int32Value(int32(data.ExternalApiSource))
	state.Aid = types.StringValue(data.Aid)
	externalApiFormFieldsElements := []ExternalAPIFieldDataSourceModel{}
	for _, element := range data.ExternalApiFormFields {
		externalApiFormFieldsElements = append(externalApiFormFieldsElements, ExternalAPIFieldDataSourceModelFrom(element))
	}
	state.ExternalApiFormFields = externalApiFormFieldsElements
	state.SharedAccountCount = types.Int32PointerValue(data.SharedAccountCount)
	state.EvtItunesProductId = types.StringValue(data.EvtItunesProductId)
	state.Name = types.StringValue(data.Name)
	state.EvtItunesBundleId = types.StringValue(data.EvtItunesBundleId)
	state.CollectAddress = types.BoolValue(data.CollectAddress)
	state.Description = types.StringValue(data.Description)
	state.TermId = types.StringValue(data.TermId)
	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
