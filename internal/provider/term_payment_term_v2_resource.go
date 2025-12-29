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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type PaymentTermV2ResourceModel struct {
	Aid                                   types.String           `tfsdk:"aid"`                                          // The application ID
	Rid                                   types.String           `tfsdk:"rid"`                                          // The resource ID
	CollectAddress                        types.Bool             `tfsdk:"collect_address"`                              // Whether to collect an address for this term
	CreateDate                            types.Int64            `tfsdk:"create_date"`                                  // The creation date
	CurrencySymbol                        types.String           `tfsdk:"currency_symbol"`                              // The currency symbol
	Description                           types.String           `tfsdk:"description"`                                  // The description of the term
	EvtVerificationPeriod                 types.Int32            `tfsdk:"evt_verification_period"`                      // The <a href = "https://docs.piano.io/external-service-term/#externaltermverification">periodicity</a> (in seconds) of checking the EVT subscription with the external service
	IsAllowedToChangeSchedulePeriodInPast types.Bool             `tfsdk:"is_allowed_to_change_schedule_period_in_past"` // Whether the term allows to change its schedule period created previously
	Name                                  types.String           `tfsdk:"name"`                                         // The term name
	PaymentAllowGift                      types.Bool             `tfsdk:"payment_allow_gift"`                           // Whether the term can be gifted
	PaymentAllowPromoCodes                types.Bool             `tfsdk:"payment_allow_promo_codes"`                    // Whether to allow promo codes to be applied
	PaymentAllowRenewDays                 types.Int32            `tfsdk:"payment_allow_renew_days"`                     // How many days in advance users user can renew
	PaymentBillingPlan                    types.String           `tfsdk:"payment_billing_plan"`                         // The billing plan for the term
	PaymentBillingPlanDescription         types.String           `tfsdk:"payment_billing_plan_description"`             // The description of the term billing plan
	PaymentCurrency                       types.String           `tfsdk:"payment_currency"`                             // The currency of the term
	PaymentFirstPrice                     types.Float64          `tfsdk:"payment_first_price"`                          // The first price of the term
	PaymentForceAutoRenew                 types.Bool             `tfsdk:"payment_force_auto_renew"`                     // Prevents users from disabling autorenewal (always "TRUE" for dynamic terms)
	PaymentHasFreeTrial                   types.Bool             `tfsdk:"payment_has_free_trial"`                       // Whether payment includes a free trial
	PaymentIsCustomPriceAvailable         types.Bool             `tfsdk:"payment_is_custom_price_available"`            // Whether users can pay more than term price
	PaymentNewCustomersOnly               types.Bool             `tfsdk:"payment_new_customers_only"`                   // Whether to show the term only to users having no dynamic or purchase conversions yet
	PaymentRenewGracePeriod               types.Int32            `tfsdk:"payment_renew_grace_period"`                   // The number of days after expiration to still allow access to the resource
	PaymentTrialNewCustomersOnly          types.Bool             `tfsdk:"payment_trial_new_customers_only"`             // Whether to allow trial period only to users having no purchases yet
	ProductCategory                       types.String           `tfsdk:"product_category"`                             // The product category
	Schedule                              *ScheduleResourceModel `tfsdk:"schedule"`
	ScheduleBilling                       types.String           `tfsdk:"schedule_billing"`      // The schedule billing
	SharedAccountCount                    types.Int32            `tfsdk:"shared_account_count"`  // The shared account count
	SharedRedemptionUrl                   types.String           `tfsdk:"shared_redemption_url"` // The shared subscription redemption URL
	TermId                                types.String           `tfsdk:"term_id"`               // The term ID
	Type                                  types.String           `tfsdk:"type"`                  // The term type
	UpdateDate                            types.Int64            `tfsdk:"update_date"`           // The update date
	VerifyOnRenewal                       types.Bool             `tfsdk:"verify_on_renewal"`     // Whether the term should be verified before renewal (if "FALSE", this step is skipped)
}

var (
	_ resource.Resource                = &PaymentTermV2Resource{}
	_ resource.ResourceWithImportState = &PaymentTermV2Resource{}
)

func NewPaymentTermV2Resource() resource.Resource {
	return &PaymentTermV2Resource{}
}

// TermDataSource defines the data source implementation.
type PaymentTermV2Resource struct {
	client *piano_publisher.Client
}

func (r *PaymentTermV2Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_payment_term_v2"
}

func (r *PaymentTermV2Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*PaymentTermV2Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Payment Term resource. Payment term is a term that is used to create a payment.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"rid": schema.StringAttribute{
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
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The term name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "The description of the term",
			},
			"payment_allow_renew_days": schema.Int32Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Computed:            true,
				Default:             int32default.StaticInt32(0),
				MarkdownDescription: "How many days in advance users user can renew",
			},
			"payment_allow_promo_codes": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to allow promo codes to be applied",
			},
			"payment_billing_plan_description": schema.StringAttribute{
				// payment_billing_plan_description is computed from paymant_billing_plan expression
				Computed:            true,
				MarkdownDescription: "The description of the term billing plan",
			},
			"is_allowed_to_change_schedule_period_in_past": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Whether the term allows to change its schedule period created previously",
			},
			"payment_is_custom_price_available": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether users can pay more than term price",
			},
			"payment_first_price": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "The first price of the term",
			},
			// https://docs.piano.io/api?endpoint=post~2F~2Fpublisher~2Fterm~2Fchange~2Foption~2Fcreate
			// change_options should be defined separately after term creation
			"payment_has_free_trial": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether payment includes a free trial",
			},
			"schedule_billing": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The schedule billing",
			},
			"collect_address": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to collect an address for this term",
			},

			"update_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The update date",
			},
			"payment_new_customers_only": schema.BoolAttribute{
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				Computed:            true,
				MarkdownDescription: "Whether to show the term only to users having no dynamic or purchase conversions yet",
			},
			"verify_on_renewal": schema.BoolAttribute{
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				Computed:            true,
				MarkdownDescription: "Whether the term should be verified before renewal (if \"FALSE\", this step is skipped)",
			},
			"create_date": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The creation date",
			},
			"evt_verification_period": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "The <a href = \"https://docs.piano.io/external-service-term/#externaltermverification\">periodicity</a> (in seconds) of checking the EVT subscription with the external service",
			},
			"payment_billing_plan": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The billing plan for the term. The value is payment billing plan expression [${CURRENCY_AMMOUNT} ${CURRENCY_UNIT}|${PERIOD_NAME}|${INTERVAL}] such as [19.99 USD|1 month|*] or [119.99 USD|12 months|1]",
			},
			"payment_allow_gift": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the term can be gifted",
			},
			"payment_force_auto_renew": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Prevents users from disabling autorenewal (always \"TRUE\" for dynamic terms)",
			},
			"schedule": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"schedule_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The schedule ID",
					},
					"name": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The schedule name",
					},
					"update_date": schema.Int64Attribute{
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The update date",
					},
					"create_date": schema.Int64Attribute{
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The creation date",
					},
					"deleted": schema.BoolAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Whether the object is deleted",
					},
					"aid": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The application ID",
					},
				},
			},
			"payment_trial_new_customers_only": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to allow trial period only to users having no purchases yet",
			},
			"payment_currency": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("USD"),
				MarkdownDescription: "The currency of the term",
			},
			"shared_account_count": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "The shared account count",
			},
			"shared_redemption_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The shared subscription redemption URL",
			},
			"currency_symbol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("$"),
				MarkdownDescription: "The currency symbol",
			},
			"product_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The product category",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The term type",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("payment", "adview", "registration", "newsletter", "external", "custom", "grant_access", "gift", "specific_email_addresses_contract", "email_domain_contract", "ip_range_contract", "dynamic", "linked"),
				},
			},
			"payment_renew_grace_period": schema.Int32Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(15),
				MarkdownDescription: "The number of days after expiration to still allow access to the resource",
			},
		},
	}
}

func (r *PaymentTermV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PaymentTermV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.PostPublisherTermPaymentCreateWithFormdataBody(ctx, piano_publisher.PostPublisherTermPaymentCreateRequest{
		Aid:                          plan.Aid.ValueString(),
		Rid:                          plan.Rid.ValueString(),
		Name:                         plan.Name.ValueString(),
		Description:                  plan.Description.ValueStringPointer(),
		PaymentBillingPlan:           plan.PaymentBillingPlan.ValueStringPointer(),
		PaymentAllowRenewDays:        plan.PaymentAllowRenewDays.ValueInt32Pointer(),
		PaymentForceAutoRenew:        plan.PaymentForceAutoRenew.ValueBoolPointer(),
		PaymentNewCustomersOnly:      plan.PaymentNewCustomersOnly.ValueBoolPointer(),
		PaymentTrialNewCustomersOnly: plan.PaymentTrialNewCustomersOnly.ValueBoolPointer(),
		PaymentAllowPromoCodes:       plan.PaymentAllowPromoCodes.ValueBoolPointer(),
		PaymentRenewGracePeriod:      plan.PaymentRenewGracePeriod.ValueInt32Pointer(),
		PaymentAllowGift:             plan.PaymentAllowGift.ValueBoolPointer(),
		SharedAccountCount:           plan.SharedAccountCount.ValueInt32Pointer(),
		SharedRedemptionUrl:          plan.SharedRedemptionUrl.ValueStringPointer(),
		CollectAddress:               plan.CollectAddress.ValueBoolPointer(),
		VerifyOnRenewal:              plan.VerifyOnRenewal.ValueBoolPointer(),
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
	tflog.Info(ctx, "created Term Payment")
	result := piano_publisher.TermResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}
	plan.TermId = types.StringValue(result.Term.TermId)
	plan.CreateDate = types.Int64Value(int64(result.Term.CreateDate))
	plan.UpdateDate = types.Int64Value(int64(result.Term.UpdateDate))
	plan.Type = types.StringValue(string(result.Term.Type))
	plan.PaymentBillingPlanDescription = types.StringValue(result.Term.PaymentBillingPlanDescription)
	plan.PaymentFirstPrice = types.Float64Value(result.Term.PaymentFirstPrice)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}
func (r *PaymentTermV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PaymentTermV2ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("%v", resp.Diagnostics))
		return
	}
	response, err := r.client.PostPublisherTermPaymentUpdateWithFormdataBody(ctx, piano_publisher.PostPublisherTermPaymentUpdateRequest{
		TermId:                       plan.TermId.ValueString(),
		Description:                  plan.Description.ValueStringPointer(),
		PaymentBillingPlan:           plan.PaymentBillingPlan.ValueStringPointer(),
		PaymentAllowRenewDays:        plan.PaymentAllowRenewDays.ValueInt32Pointer(),
		PaymentForceAutoRenew:        plan.PaymentForceAutoRenew.ValueBoolPointer(),
		PaymentNewCustomersOnly:      plan.PaymentNewCustomersOnly.ValueBoolPointer(),
		PaymentTrialNewCustomersOnly: plan.PaymentTrialNewCustomersOnly.ValueBoolPointer(),
		PaymentAllowPromoCodes:       plan.PaymentAllowPromoCodes.ValueBoolPointer(),
		PaymentRenewGracePeriod:      plan.PaymentRenewGracePeriod.ValueInt32Pointer(),
		PaymentAllowGift:             plan.PaymentAllowGift.ValueBoolPointer(),
		SharedAccountCount:           plan.SharedAccountCount.ValueInt32Pointer(),
		SharedRedemptionUrl:          plan.SharedRedemptionUrl.ValueStringPointer(),
		CollectAddress:               plan.CollectAddress.ValueBoolPointer(),
		VerifyOnRenewal:              plan.VerifyOnRenewal.ValueBoolPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update resource, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	tflog.Info(ctx, "update Payment term")
	result := piano_publisher.TermResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	plan.UpdateDate = types.Int64Value(int64(result.Term.UpdateDate))
	plan.PaymentBillingPlanDescription = types.StringValue(result.Term.PaymentBillingPlanDescription)
	plan.PaymentFirstPrice = types.Float64Value(result.Term.PaymentFirstPrice)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PaymentTermV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PaymentTermV2ResourceModel

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

	state.PaymentRenewGracePeriod = types.Int32Value(data.PaymentRenewGracePeriod)

	state.Type = types.StringValue(string(data.Type))
	if data.ProductCategory != "" {
		state.ProductCategory = types.StringValue(data.ProductCategory)
	}
	state.CurrencySymbol = types.StringValue(data.CurrencySymbol)
	state.SharedAccountCount = types.Int32PointerValue(data.SharedAccountCount)
	state.SharedRedemptionUrl = types.StringPointerValue(data.SharedRedemptionUrl)
	state.PaymentCurrency = types.StringValue(data.PaymentCurrency)
	state.PaymentTrialNewCustomersOnly = types.BoolValue(data.PaymentTrialNewCustomersOnly)
	if data.Schedule != nil {
		Schedule := ScheduleResourceModelFrom(*data.Schedule)
		state.Schedule = &Schedule
	}
	state.PaymentForceAutoRenew = types.BoolValue(data.PaymentForceAutoRenew)
	state.PaymentAllowGift = types.BoolValue(data.PaymentAllowGift)
	state.PaymentBillingPlan = types.StringValue(data.PaymentBillingPlan)

	Resource := ResourceResourceModelFrom(data.Resource)
	state.Rid = Resource.Rid
	state.EvtVerificationPeriod = types.Int32PointerValue(data.EvtVerificationPeriod)
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.VerifyOnRenewal = types.BoolValue(data.VerifyOnRenewal)
	state.PaymentNewCustomersOnly = types.BoolValue(data.PaymentNewCustomersOnly)
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	state.CollectAddress = types.BoolValue(data.CollectAddress)
	state.ScheduleBilling = types.StringPointerValue(data.ScheduleBilling)
	state.PaymentHasFreeTrial = types.BoolValue(data.PaymentHasFreeTrial)
	state.Aid = types.StringValue(data.Aid)

	state.PaymentFirstPrice = types.Float64Value(data.PaymentFirstPrice)
	state.Name = types.StringValue(data.Name)

	state.TermId = types.StringValue(data.TermId)
	state.PaymentIsCustomPriceAvailable = types.BoolValue(data.PaymentIsCustomPriceAvailable)
	state.IsAllowedToChangeSchedulePeriodInPast = types.BoolValue(data.IsAllowedToChangeSchedulePeriodInPast)

	state.PaymentBillingPlanDescription = types.StringValue(data.PaymentBillingPlanDescription)
	state.PaymentAllowPromoCodes = types.BoolValue(data.PaymentAllowPromoCodes)
	state.Description = types.StringValue(data.Description)
	state.PaymentAllowRenewDays = types.Int32Value(data.PaymentAllowRenewDays)

	tflog.Trace(ctx, "read a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PaymentTermV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PaymentTermV2ResourceModel
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
	_, err = syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}
}
func (r *PaymentTermV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := TermResourceIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Term resource id", fmt.Sprintf("Unable to parse contract resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), id.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("term_id"), id.TermId)...)
}
