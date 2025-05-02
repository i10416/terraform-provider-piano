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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ExternalAPIFieldResourceModel struct {
	DefaultValue types.String `tfsdk:"default_value"` // Default value for the field. It will be pre-entered on the form
	Description  types.String `tfsdk:"description"`   // The field description, some information about what information should be entered
	Editable     types.String `tfsdk:"editable"`      // Whether the object is editable
	FieldName    types.String `tfsdk:"field_name"`    // The name of the field to be used to submit to the external system
	FieldTitle   types.String `tfsdk:"field_title"`   // The title of the field to be displayed to the user
	Hidden       types.Bool   `tfsdk:"hidden"`        // Whether the field will be submitted hiddenly from the user, default value is required
	Mandatory    types.Bool   `tfsdk:"mandatory"`     // Whether the field is required
	Order        types.Int32  `tfsdk:"order"`         // Field order in the list
	Type         types.String `tfsdk:"type"`          // Field type
}

type ExternalTermResourceModel struct {
	Aid                      types.String `tfsdk:"aid"`                          // The application ID
	TermId                   types.String `tfsdk:"term_id"`                      // The term ID
	ExternalApiId            types.String `tfsdk:"external_api_id"`              // The ID of the external API configuration
	Name                     types.String `tfsdk:"name"`                         // The term name
	Description              types.String `tfsdk:"description"`                  // The description of the term
	EvtFixedTimeAccessPeriod types.Int32  `tfsdk:"evt_fixed_time_access_period"` // The period to grant access for (in days)
	EvtGooglePlayProductId   types.String `tfsdk:"evt_google_play_product_id"`   // Google Play's product ID
	EvtGracePeriod           types.Int32  `tfsdk:"evt_grace_period"`             // The External API grace period
	EvtItunesBundleId        types.String `tfsdk:"evt_itunes_bundle_id"`         // iTunes's bundle ID
	EvtItunesProductId       types.String `tfsdk:"evt_itunes_product_id"`        // iTunes's product ID
	EvtVerificationPeriod    types.Int32  `tfsdk:"evt_verification_period"`      // The <a href = "https://docs.piano.io/external-service-term/#externaltermverification">periodicity</a> (in seconds) of checking the EVT subscription with the external service
	SharedAccountCount       types.Int32  `tfsdk:"shared_account_count"`         // The count of allowed shared-subscription accounts
	SharedRedemptionUrl      types.String `tfsdk:"shared_redemption_url"`        // The shared subscription redemption URL
	// read only
	ExternalApiName       types.String                           `tfsdk:"external_api_name"`   // The name of the external API configuration
	ExternalApiSource     types.Int32                            `tfsdk:"external_api_source"` // The source of the external API configuration
	CreateDate            types.Int64                            `tfsdk:"create_date"`         // The creation date
	UpdateDate            types.Int64                            `tfsdk:"update_date"`         // The update date
	Type                  types.String                           `tfsdk:"type"`                // The term type
	Resource              *ResourceResourceModel                 `tfsdk:"resource"`
	ExternalApiFormFields ExternalAPIFieldResourceModelListValue `tfsdk:"external_api_form_fields"`
}

func (r *ExternalTermResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_term"
}
func (*ExternalTermResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ExternalTerm resource. External term is a term that is created by the external API.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"term_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The term ID",
			},
			"external_api_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the external API configuration",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The term name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description of the term",
			},
			"evt_itunes_bundle_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "iTunes's bundle ID",
			},
			"evt_itunes_product_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "iTunes's product ID",
			},
			"shared_account_count": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "The count of allowed shared-subscription accounts",
			},
			"evt_verification_period": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "The <a href = \"https://docs.piano.io/external-service-term/#externaltermverification\">periodicity</a> (in seconds) of checking the EVT subscription with the external service",
			},
			"evt_google_play_product_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Google Play's product ID",
			},
			"evt_fixed_time_access_period": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "The period to grant access for (in days)",
			},
			"shared_redemption_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The shared subscription redemption URL",
			},
			"evt_grace_period": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "The External API grace period",
			},
			"external_api_form_fields": schema.ListNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				CustomType: ExternalAPIFieldResourceModelList{
					ListType: basetypes.ListType{
						ElemType: ExternalAPIFieldAttrType(),
					},
				},
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
			"resource": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"rid": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The resource ID",
					},
					"aid": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The application ID",
					},
					"bundle_type": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The resource bundle type",
						Validators: []validator.String{
							stringvalidator.OneOf("undefined", "fixed", "tagged", "fixed_v2"),
						},
					},
					"image_url": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The URL of the resource image",
					},
					"purchase_url": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The URL of the purchase page",
					},

					"description": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The resource description",
					},
					"update_date": schema.Int64Attribute{
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The update date",
					},
					"type": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The type of the resource (0: Standard, 4: Bundle)",
						Validators: []validator.String{
							stringvalidator.OneOf("standard", "bundle", "print"),
						},
					},
					"deleted": schema.BoolAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Whether the object is deleted",
					},

					"name": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The name",
					},
					"create_date": schema.Int64Attribute{
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The creation date",
					},
					"is_fbia_resource": schema.BoolAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Enable the resource for Facebook Subscriptions in Instant Articles",
					},
					"external_id": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The external ID; defined by the client",
					},
					"publish_date": schema.Int64Attribute{
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The publish date",
					},
					"resource_url": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The URL of the resource",
					},
					"disabled": schema.BoolAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Whether the object is disabled",
					},
				},
			},
			"external_api_source": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The source of the external API configuration",
			},
			"external_api_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The name of the external API configuration",
			},
			"update_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The update date",
			},
			"create_date": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The creation date",
			},
			"type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The term type",
				Validators: []validator.String{
					stringvalidator.OneOf("payment", "adview", "registration", "newsletter", "external", "custom", "grant_access", "gift", "specific_email_addresses_contract", "email_domain_contract", "ip_range_contract", "dynamic", "linked"),
				},
			},
		},
	}
}
func NewExternalTermResource() resource.Resource {
	return &ExternalTermResource{}
}

// ExternalTermResource defines the data source implementation.
type ExternalTermResource struct {
	client *piano_publisher.Client
}

func (r *ExternalTermResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExternalTermResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state ExternalTermResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("%v", resp.Diagnostics))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("creating resource %s in %s", state.Name.ValueString(), state.Aid.ValueString()))

	response, err := r.client.PostPublisherTermExternalCreateWithFormdataBody(ctx, piano_publisher.PostPublisherTermExternalCreateFormdataRequestBody{
		Aid:                      state.Aid.ValueString(),
		Rid:                      state.Resource.Rid.ValueString(),
		ExternalApiId:            state.ExternalApiId.ValueString(),
		Name:                     state.Name.ValueString(),
		Description:              state.Description.ValueStringPointer(),
		EvtFixedTimeAccessPeriod: state.EvtFixedTimeAccessPeriod.ValueInt32Pointer(),
		EvtGracePeriod:           state.EvtGracePeriod.ValueInt32Pointer(),
		EvtVerificationPeriod:    state.EvtVerificationPeriod.ValueInt32Pointer(),
		EvtItunesBundleId:        state.EvtItunesBundleId.ValueStringPointer(),
		EvtItunesProductId:       state.EvtItunesProductId.ValueStringPointer(),
		EvtGooglePlayProductId:   state.EvtGooglePlayProductId.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
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
	state.EvtGracePeriod = types.Int32Value(data.EvtGracePeriod)
	state.EvtFixedTimeAccessPeriod = types.Int32PointerValue(data.EvtFixedTimeAccessPeriod)
	Resource := ResourceResourceModelFrom(data.Resource)
	state.Resource = &Resource
	state.EvtGooglePlayProductId = types.StringPointerValue(data.EvtGooglePlayProductId)
	state.EvtVerificationPeriod = types.Int32PointerValue(data.EvtVerificationPeriod)
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	state.ExternalApiName = types.StringValue(data.ExternalApiName)
	state.ExternalApiSource = types.Int32Value(int32(data.ExternalApiSource))
	state.Aid = types.StringValue(data.Aid)
	state.SharedAccountCount = types.Int32PointerValue(data.SharedAccountCount)
	state.EvtItunesProductId = types.StringValue(data.EvtItunesProductId)
	state.Name = types.StringValue(data.Name)

	externalApiFormFieldsElements := []ExternalAPIFieldResourceModel{}
	for _, element := range data.ExternalApiFormFields {
		externalApiFormFieldsElements = append(externalApiFormFieldsElements, ExternalAPIFieldResourceModelFrom(element))
	}
	listValue, diags := basetypes.NewListValueFrom(ctx, ExternalAPIFieldAttrType(), externalApiFormFieldsElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	state.ExternalApiFormFields = ExternalAPIFieldResourceModelListValue{ListValue: listValue}
	state.EvtItunesBundleId = types.StringValue(data.EvtItunesBundleId)
	state.Description = types.StringValue(data.Description)
	state.TermId = types.StringValue(data.TermId)
	tflog.Info(ctx, fmt.Sprintf("complete creating resource %s(id: %s)", state.Name, state.TermId))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *ExternalTermResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state ExternalTermResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("%v", resp.Diagnostics))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("creating resource %s in %s", state.Name.ValueString(), state.Aid.ValueString()))
	request := piano_publisher.PostPublisherTermExternalUpdateFormdataRequestBody{
		TermId:                   state.TermId.ValueString(),
		ExternalApiId:            state.ExternalApiId.ValueString(),
		Name:                     state.Name.ValueString(),
		Description:              state.Description.ValueStringPointer(),
		EvtFixedTimeAccessPeriod: state.EvtFixedTimeAccessPeriod.ValueInt32Pointer(),
		EvtGracePeriod:           state.EvtGracePeriod.ValueInt32Pointer(),
		EvtVerificationPeriod:    state.EvtVerificationPeriod.ValueInt32Pointer(),
		EvtItunesBundleId:        state.EvtItunesBundleId.ValueStringPointer(),
		EvtItunesProductId:       state.EvtItunesProductId.ValueStringPointer(),
		EvtGooglePlayProductId:   state.EvtGooglePlayProductId.ValueStringPointer(),
		SharedAccountCount:       state.SharedAccountCount.ValueInt32Pointer(),
		SharedRedemptionUrl:      state.SharedRedemptionUrl.ValueStringPointer(),
	}
	if !state.Resource.Rid.IsNull() {
		request.Rid = state.Resource.Rid.ValueStringPointer()
	}
	response, err := r.client.PostPublisherTermExternalUpdateWithFormdataBody(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
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
	state.EvtGracePeriod = types.Int32Value(data.EvtGracePeriod)
	state.EvtFixedTimeAccessPeriod = types.Int32PointerValue(data.EvtFixedTimeAccessPeriod)
	Resource := ResourceResourceModelFrom(data.Resource)
	state.Resource = &Resource
	state.EvtGooglePlayProductId = types.StringPointerValue(data.EvtGooglePlayProductId)
	state.EvtVerificationPeriod = types.Int32PointerValue(data.EvtVerificationPeriod)
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	state.ExternalApiName = types.StringValue(data.ExternalApiName)
	state.ExternalApiSource = types.Int32Value(int32(data.ExternalApiSource))
	state.Aid = types.StringValue(data.Aid)

	externalApiFormFieldsElements := []ExternalAPIFieldResourceModel{}
	for _, element := range data.ExternalApiFormFields {
		externalApiFormFieldsElements = append(externalApiFormFieldsElements, ExternalAPIFieldResourceModelFrom(element))
	}
	listValue, diags := basetypes.NewListValueFrom(ctx, ExternalAPIFieldAttrType(), externalApiFormFieldsElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	state.ExternalApiFormFields = ExternalAPIFieldResourceModelListValue{ListValue: listValue}
	state.SharedAccountCount = types.Int32PointerValue(data.SharedAccountCount)
	state.EvtItunesProductId = types.StringValue(data.EvtItunesProductId)
	state.Name = types.StringValue(data.Name)
	state.EvtItunesBundleId = types.StringValue(data.EvtItunesBundleId)
	state.Description = types.StringValue(data.Description)
	tflog.Info(ctx, fmt.Sprintf("complete updating resource %s(id: %s)", state.Name, state.TermId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ExternalTermResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ExternalTermResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

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
	state.EvtGracePeriod = types.Int32Value(data.EvtGracePeriod)
	state.EvtFixedTimeAccessPeriod = types.Int32PointerValue(data.EvtFixedTimeAccessPeriod)
	Resource := ResourceResourceModelFrom(data.Resource)
	state.Resource = &Resource
	state.EvtGooglePlayProductId = types.StringPointerValue(data.EvtGooglePlayProductId)
	state.EvtVerificationPeriod = types.Int32PointerValue(data.EvtVerificationPeriod)
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	state.ExternalApiName = types.StringValue(data.ExternalApiName)
	state.ExternalApiSource = types.Int32Value(int32(data.ExternalApiSource))
	state.Aid = types.StringValue(data.Aid)

	externalApiFormFieldsElements := []ExternalAPIFieldResourceModel{}
	for _, element := range data.ExternalApiFormFields {
		externalApiFormFieldsElements = append(externalApiFormFieldsElements, ExternalAPIFieldResourceModelFrom(element))
	}
	listValue, diags := basetypes.NewListValueFrom(ctx, ExternalAPIFieldAttrType(), externalApiFormFieldsElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	state.ExternalApiFormFields = ExternalAPIFieldResourceModelListValue{ListValue: listValue}
	state.SharedAccountCount = types.Int32PointerValue(data.SharedAccountCount)
	state.EvtItunesProductId = types.StringValue(data.EvtItunesProductId)
	state.Name = types.StringValue(data.Name)
	state.EvtItunesBundleId = types.StringValue(data.EvtItunesBundleId)
	state.Description = types.StringValue(data.Description)
	tflog.Trace(ctx, "read a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ExternalTermResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ExternalTermResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("deleting Term %s:%s in $%s", state.Name.ValueString(), state.TermId.ValueString(), state.Aid.ValueString()))
	response, err := r.client.PostPublisherTermDeleteWithFormdataBody(ctx, piano_publisher.PostPublisherTermDeleteFormdataRequestBody{
		TermId: state.TermId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete resource, got error: %s", err))
		return
	}
	_, err = syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}
}

func (r *ExternalTermResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := TermResourceIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Term resource id", fmt.Sprintf("Unable to parse contract resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), id.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("term_id"), id.TermId)...)
}

func ExternalAPIFieldResourceModelFrom(data piano_publisher.ExternalAPIField) ExternalAPIFieldResourceModel {
	ret := ExternalAPIFieldResourceModel{}
	ret.Editable = types.StringValue(data.Editable)
	ret.FieldName = types.StringValue(data.FieldName)
	ret.Type = types.StringValue(string(data.Type))
	ret.FieldTitle = types.StringValue(data.FieldTitle)
	ret.DefaultValue = types.StringPointerValue(data.DefaultValue)
	ret.Order = types.Int32Value(data.Order)
	ret.Hidden = types.BoolValue(data.Hidden)
	ret.Description = types.StringValue(data.Description)
	ret.Mandatory = types.BoolValue(data.Mandatory)
	return ret
}
