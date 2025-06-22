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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type PeriodResourceModel struct {
	BeginDate     types.Int64  `tfsdk:"begin_date"`      // The date when the period begins
	CreateDate    types.Int64  `tfsdk:"create_date"`     // The creation date
	Deleted       types.Bool   `tfsdk:"deleted"`         // Whether the object is deleted
	EndDate       types.Int64  `tfsdk:"end_date"`        // The date when the period ends
	IsActive      types.Bool   `tfsdk:"is_active"`       // Whether the period is active. A period is in the Active state when the sell date is passed but the end date is not reached
	IsSaleStarted types.Bool   `tfsdk:"is_sale_started"` // Whether sale is started for the period
	Name          types.String `tfsdk:"name"`            // The period name
	PeriodId      types.String `tfsdk:"period_id"`       // The period ID
	SellDate      types.Int64  `tfsdk:"sell_date"`       // The sell date of the period
	UpdateDate    types.Int64  `tfsdk:"update_date"`     // The update date
}

type AdvancedOptionsResourceModel struct {
	ShowOptions []types.String `tfsdk:"show_options"`
}
type TermBriefResourceModel struct {
	Disabled types.Bool   `tfsdk:"disabled"` // Whether the term is disabled
	Name     types.String `tfsdk:"name"`     // The term name
	TermId   types.String `tfsdk:"term_id"`  // The term ID
}

type PaymentTermResourceModel struct {
	Aid                                   types.String                    `tfsdk:"aid"`            // The application ID
	BillingConfig                         types.String                    `tfsdk:"billing_config"` // The type of billing config
	ChangeOptions                         []TermChangeOptionResourceModel `tfsdk:"change_options"`
	CollectAddress                        types.Bool                      `tfsdk:"collect_address"`                              // Whether to collect an address for this term
	CreateDate                            types.Int64                     `tfsdk:"create_date"`                                  // The creation date
	CurrencySymbol                        types.String                    `tfsdk:"currency_symbol"`                              // The currency symbol
	Description                           types.String                    `tfsdk:"description"`                                  // The description of the term
	EvtVerificationPeriod                 types.Int32                     `tfsdk:"evt_verification_period"`                      // The <a href = "https://docs.piano.io/external-service-term/#externaltermverification">periodicity</a> (in seconds) of checking the EVT subscription with the external service
	IsAllowedToChangeSchedulePeriodInPast types.Bool                      `tfsdk:"is_allowed_to_change_schedule_period_in_past"` // Whether the term allows to change its schedule period created previously
	Name                                  types.String                    `tfsdk:"name"`                                         // The term name
	PaymentAllowGift                      types.Bool                      `tfsdk:"payment_allow_gift"`                           // Whether the term can be gifted
	PaymentAllowPromoCodes                types.Bool                      `tfsdk:"payment_allow_promo_codes"`                    // Whether to allow promo codes to be applied
	PaymentAllowRenewDays                 types.Int32                     `tfsdk:"payment_allow_renew_days"`                     // How many days in advance users user can renew
	PaymentBillingPlan                    types.String                    `tfsdk:"payment_billing_plan"`                         // The billing plan for the term
	PaymentBillingPlanDescription         types.String                    `tfsdk:"payment_billing_plan_description"`             // The description of the term billing plan
	PaymentCurrency                       types.String                    `tfsdk:"payment_currency"`                             // The currency of the term
	PaymentFirstPrice                     types.Float64                   `tfsdk:"payment_first_price"`                          // The first price of the term
	PaymentForceAutoRenew                 types.Bool                      `tfsdk:"payment_force_auto_renew"`                     // Prevents users from disabling autorenewal (always "TRUE" for dynamic terms)
	PaymentHasFreeTrial                   types.Bool                      `tfsdk:"payment_has_free_trial"`                       // Whether payment includes a free trial
	PaymentIsCustomPriceAvailable         types.Bool                      `tfsdk:"payment_is_custom_price_available"`            // Whether users can pay more than term price
	PaymentIsSubscription                 types.Bool                      `tfsdk:"payment_is_subscription"`                      // Whether this term (payment or dynamic) is a subscription (unlike one-off)
	PaymentNewCustomersOnly               types.Bool                      `tfsdk:"payment_new_customers_only"`                   // Whether to show the term only to users having no dynamic or purchase conversions yet
	PaymentRenewGracePeriod               types.Int32                     `tfsdk:"payment_renew_grace_period"`                   // The number of days after expiration to still allow access to the resource
	PaymentTrialNewCustomersOnly          types.Bool                      `tfsdk:"payment_trial_new_customers_only"`             // Whether to allow trial period only to users having no purchases yet
	ProductCategory                       types.String                    `tfsdk:"product_category"`                             // The product category
	Resource                              *ResourceResourceModel          `tfsdk:"resource"`
	Schedule                              *ScheduleResourceModel          `tfsdk:"schedule"`
	ScheduleBilling                       types.String                    `tfsdk:"schedule_billing"`        // The schedule billing
	SharedRedemptionUrl                   types.String                    `tfsdk:"shared_redemption_url"`   // The shared subscription redemption URL
	TermBillingDescriptor                 types.String                    `tfsdk:"term_billing_descriptor"` // The term billing descriptor
	TermId                                types.String                    `tfsdk:"term_id"`                 // The term ID
	Type                                  types.String                    `tfsdk:"type"`                    // The term type
	UpdateDate                            types.Int64                     `tfsdk:"update_date"`             // The update date
	VerifyOnRenewal                       types.Bool                      `tfsdk:"verify_on_renewal"`       // Whether the term should be verified before renewal (if "FALSE", this step is skipped)
}

type TermChangeOptionResourceModel struct {
	AdvancedOptions    *AdvancedOptionsResourceModel `tfsdk:"advanced_options"`
	BillingTiming      types.String                  `tfsdk:"billing_timing"`        // The billing timing(0: immediate term change;1: term change at the end of the current cycle;2: term change on the next sell date;3: term change at the end of the current period)
	CollectAddress     types.Bool                    `tfsdk:"collect_address"`       // Whether to collect an address for this term
	Description        types.String                  `tfsdk:"description"`           // A description of the term change option; provided by the client
	FromBillingPlan    types.String                  `tfsdk:"from_billing_plan"`     // The "From" billing plan
	FromPeriodId       types.String                  `tfsdk:"from_period_id"`        // The ID of the "From" term period
	FromPeriodName     types.String                  `tfsdk:"from_period_name"`      // The name of the "From" term period
	FromResourceId     types.String                  `tfsdk:"from_resource_id"`      // The ID of the "From" resource
	FromResourceName   types.String                  `tfsdk:"from_resource_name"`    // The name of the "From" resource
	FromScheduled      types.Bool                    `tfsdk:"from_scheduled"`        // Whether the subscription is upgraded from a scheduled term
	FromTermId         types.String                  `tfsdk:"from_term_id"`          // The ID of the "From" term
	FromTermName       types.String                  `tfsdk:"from_term_name"`        // The name of the "From" term
	ImmediateAccess    types.Bool                    `tfsdk:"immediate_access"`      // Whether the access begins immediately
	IncludeTrial       types.Bool                    `tfsdk:"include_trial"`         // Whether trial is enabled (not in use, always "FALSE")
	ProrateAccess      types.Bool                    `tfsdk:"prorate_access"`        // Whether the <a href="https://docs.piano.io/upgrades/?paragraphId=b27954ef84407e4#prorate-billing-amount">Prorate billing amount</a> function is enabled
	SharedAccountCount types.Int32                   `tfsdk:"shared_account_count"`  // The count of allowed shared-subscription accounts
	TermChangeOptionId types.String                  `tfsdk:"term_change_option_id"` // The ID of the term change option
	ToBillingPlan      types.String                  `tfsdk:"to_billing_plan"`       // The "To" billing plan
	ToPeriodId         types.String                  `tfsdk:"to_period_id"`          // The ID of the "To" term period
	ToPeriodName       types.String                  `tfsdk:"to_period_name"`        // The period name of the "To" term
	ToResourceId       types.String                  `tfsdk:"to_resource_id"`        // The ID of the "To" resource
	ToResourceName     types.String                  `tfsdk:"to_resource_name"`      // The name of the "To" resource
	ToScheduled        types.Bool                    `tfsdk:"to_scheduled"`          // Whether the subscription is upgraded to a scheduled term
	ToTermId           types.String                  `tfsdk:"to_term_id"`            // The ID of the "To" term
	ToTermName         types.String                  `tfsdk:"to_term_name"`          // The name of the "To" term
}

type PaymentBillingPlanTableResourceModel struct {
	Billing                types.String  `tfsdk:"billing"` // payment condition such as "one payment of $99.99" or "$119.99 per year"
	BillingInfo            types.String  `tfsdk:"billing_info"`
	BillingPeriod          types.String  `tfsdk:"billing_period"`
	Currency               types.String  `tfsdk:"currency"`
	Cycles                 types.String  `tfsdk:"cycles"`
	Date                   types.String  `tfsdk:"date"`       // Payment billing plan table date for humans such as "Today" or "Apr 17, 2026"
	DateValue              types.Int64   `tfsdk:"date_value"` // Payment billing plan table date in timestamp
	Duration               types.String  `tfsdk:"duration"`
	IsFree                 types.String  `tfsdk:"is_free"`
	IsFreeTrial            types.String  `tfsdk:"is_free_trial"`
	IsPayWhatYouWant       types.String  `tfsdk:"is_pay_what_you_want"`
	IsTrial                types.String  `tfsdk:"is_trial"`
	Period                 types.String  `tfsdk:"period"`
	Price                  types.String  `tfsdk:"price"` // price with currency unit symbol
	PriceAndTax            types.Float64 `tfsdk:"price_and_tax"`
	PriceAndTaxInMinorUnit types.Float32 `tfsdk:"price_and_tax_in_minor_unit"`
	PriceChargedStr        types.String  `tfsdk:"price_charged_str"` // price with currency unit symbol
	PriceValue             types.Float64 `tfsdk:"price_value"`
	ShortPeriod            types.String  `tfsdk:"short_period"` // human readable billing period in shorter expression such as /yr
	TotalBilling           types.String  `tfsdk:"total_billing"`
}

var (
	_ resource.Resource                = &PaymentTermResource{}
	_ resource.ResourceWithImportState = &PaymentTermResource{}
)

func NewPaymentTermResource() resource.Resource {
	return &PaymentTermResource{}
}

// TermDataSource defines the data source implementation.
type PaymentTermResource struct {
	client *piano_publisher.Client
}

func (r *PaymentTermResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_payment_term"
}

func (r *PaymentTermResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*PaymentTermResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Payment Term resource. Payment term is a term that is used to create a payment.",
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
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "How many days in advance users user can renew",
			},
			"payment_allow_promo_codes": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether to allow promo codes to be applied",
			},
			"payment_billing_plan_description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description of the term billing plan",
			},
			"is_allowed_to_change_schedule_period_in_past": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the term allows to change its schedule period created previously",
			},
			"payment_is_custom_price_available": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether users can pay more than term price",
			},
			"payment_is_subscription": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether this term (payment or dynamic) is a subscription (unlike one-off)",
			},
			"payment_first_price": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             float64default.StaticFloat64(0),
				MarkdownDescription: "The first price of the term",
			},
			"change_options": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"from_resource_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the \"From\" resource",
						},
						"from_period_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the \"From\" term period",
						},
						"to_period_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The period name of the \"To\" term",
						},
						"prorate_access": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the <a href=\"https://docs.piano.io/upgrades/?paragraphId=b27954ef84407e4#prorate-billing-amount\">Prorate billing amount</a> function is enabled",
						},
						"to_billing_plan": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The \"To\" billing plan",
						},
						"from_period_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the \"From\" term period",
						},
						"from_term_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the \"From\" term",
						},
						"advanced_options": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"show_options": schema.ListAttribute{
									Computed:    true,
									ElementType: basetypes.StringType{},
								},
							},
						},
						"term_change_option_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the term change option",
						},
						"to_term_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the \"To\" term",
						},
						"to_term_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the \"To\" term",
						},
						"include_trial": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether trial is enabled (not in use, always \"FALSE\")",
						},
						"immediate_access": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the access begins immediately",
						},
						"from_term_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the \"From\" term",
						},
						"from_billing_plan": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The \"From\" billing plan",
						},
						"shared_account_count": schema.Int32Attribute{
							Computed:            true,
							MarkdownDescription: "The count of allowed shared-subscription accounts",
						},
						"from_resource_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the \"From\" resource",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A description of the term change option; provided by the client",
						},
						"from_scheduled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the subscription is upgraded from a scheduled term",
						},
						"billing_timing": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The billing timing(0: immediate term change;1: term change at the end of the current cycle;2: term change on the next sell date;3: term change at the end of the current period)",
							Validators: []validator.String{
								stringvalidator.OneOf("0", "1", "2", "3"),
							},
						},
						"collect_address": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Whether to collect an address for this term",
						},
						"to_resource_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the \"To\" resource",
						},
						"to_period_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the \"To\" term period",
						},
						"to_resource_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the \"To\" resource",
						},
						"to_scheduled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the subscription is upgraded to a scheduled term",
						},
					},
				},
			},
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"term_billing_descriptor": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "The term billing descriptor",
			},
			"payment_new_customers_only": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether to show the term only to users having no dynamic or purchase conversions yet",
			},
			"billing_config": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "The type of billing config",
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
				Computed:            true,
				MarkdownDescription: "The <a href = \"https://docs.piano.io/external-service-term/#externaltermverification\">periodicity</a> (in seconds) of checking the EVT subscription with the external service",
			},
			"resource": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"aid": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The application ID",
					},
					"rid": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "The resource ID",
					},
					"bundle_type": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "The resource bundle type",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
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
			"payment_billing_plan": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The billing plan for the term",
			},
			"payment_allow_gift": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the term can be gifted",
			},
			"payment_force_auto_renew": schema.BoolAttribute{
				Required:            true,
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
				Required:            true,
				MarkdownDescription: "Whether to allow trial period only to users having no purchases yet",
			},
			"payment_currency": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "The currency of the term",
			},
			"shared_redemption_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The shared subscription redemption URL",
			},
			"currency_symbol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "The currency symbol",
			},
			"product_category": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
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
				Default:             int32default.StaticInt32(0),
				MarkdownDescription: "The number of days after expiration to still allow access to the resource",
			},
		},
	}
}

func PeriodResourceModelFrom(data piano_publisher.Period) PeriodResourceModel {
	ret := PeriodResourceModel{}
	ret.IsActive = types.BoolValue(data.IsActive)
	ret.Name = types.StringValue(data.Name)
	ret.Deleted = types.BoolValue(data.Deleted)
	ret.BeginDate = types.Int64Value(int64(data.BeginDate))
	ret.EndDate = types.Int64Value(int64(data.EndDate))
	ret.CreateDate = types.Int64Value(int64(data.CreateDate))
	ret.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	ret.SellDate = types.Int64Value(int64(data.SellDate))
	ret.PeriodId = types.StringValue(data.PeriodId)
	ret.IsSaleStarted = types.BoolValue(data.IsSaleStarted)
	return ret
}
func LightOfferResourceModelFrom(data piano_publisher.LightOffer) LightOfferResourceModel {
	ret := LightOfferResourceModel{}
	ret.Name = types.StringValue(data.Name)
	ret.OfferId = types.StringValue(data.OfferId)
	return ret
}
func AdvancedOptionsResourceModelFrom(data piano_publisher.AdvancedOptions) AdvancedOptionsResourceModel {
	ret := AdvancedOptionsResourceModel{}
	showOptionsElements := []types.String{}
	for _, element := range data.ShowOptions {
		showOptionsElements = append(showOptionsElements, types.StringValue(element))
	}
	ret.ShowOptions = showOptionsElements
	return ret
}
func TermChangeOptionResourceModelFrom(data piano_publisher.TermChangeOption) TermChangeOptionResourceModel {
	ret := TermChangeOptionResourceModel{}
	ret.ToScheduled = types.BoolValue(data.ToScheduled)
	ret.ToResourceName = types.StringValue(data.ToResourceName)
	ret.ToPeriodId = types.StringPointerValue(data.ToPeriodId)
	ret.ToResourceId = types.StringValue(data.ToResourceId)
	ret.CollectAddress = types.BoolValue(data.CollectAddress)
	ret.BillingTiming = types.StringValue(string(data.BillingTiming))
	ret.FromScheduled = types.BoolValue(data.FromScheduled)
	ret.Description = types.StringValue(data.Description)
	ret.FromResourceName = types.StringValue(data.FromResourceName)
	ret.SharedAccountCount = types.Int32PointerValue(data.SharedAccountCount)
	ret.FromBillingPlan = types.StringValue(data.FromBillingPlan)
	ret.FromTermId = types.StringValue(data.FromTermId)
	ret.ImmediateAccess = types.BoolValue(data.ImmediateAccess)
	ret.IncludeTrial = types.BoolValue(data.IncludeTrial)
	ret.ToTermId = types.StringValue(data.ToTermId)
	ret.ToTermName = types.StringValue(data.ToTermName)
	ret.TermChangeOptionId = types.StringValue(data.TermChangeOptionId)
	AdvancedOptions := AdvancedOptionsResourceModelFrom(data.AdvancedOptions)
	ret.AdvancedOptions = &AdvancedOptions
	ret.FromTermName = types.StringValue(data.FromTermName)
	ret.FromPeriodName = types.StringPointerValue(data.FromPeriodName)
	ret.ToBillingPlan = types.StringValue(data.ToBillingPlan)
	ret.ProrateAccess = types.BoolValue(data.ProrateAccess)
	ret.ToPeriodName = types.StringPointerValue(data.ToPeriodName)
	ret.FromPeriodId = types.StringPointerValue(data.FromPeriodId)
	ret.FromResourceId = types.StringValue(data.FromResourceId)
	return ret
}
func ScheduleResourceModelFrom(data piano_publisher.Schedule) ScheduleResourceModel {
	ret := ScheduleResourceModel{}

	ret.Aid = types.StringValue(data.Aid)
	ret.Deleted = types.BoolValue(data.Deleted)
	ret.ScheduleId = types.StringValue(data.ScheduleId)
	ret.CreateDate = types.Int64Value(int64(data.CreateDate))
	ret.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	ret.Name = types.StringValue(data.Name)
	return ret
}

func PaymentBillingPlanTableResourceModelFrom(data piano_publisher.PaymentBillingPlanTable) PaymentBillingPlanTableResourceModel {
	ret := PaymentBillingPlanTableResourceModel{}
	ret.IsTrial = types.StringPointerValue(data.IsTrial)
	ret.Cycles = types.StringPointerValue(data.Cycles)
	ret.PriceValue = types.Float64PointerValue(data.PriceValue)
	ret.PriceAndTaxInMinorUnit = types.Float32PointerValue(data.PriceAndTaxInMinorUnit)
	ret.ShortPeriod = types.StringPointerValue(data.ShortPeriod)
	ret.IsPayWhatYouWant = types.StringPointerValue(data.IsPayWhatYouWant)
	ret.TotalBilling = types.StringPointerValue(data.TotalBilling)
	ret.Currency = types.StringPointerValue(data.Currency)
	ret.Period = types.StringPointerValue(data.Period)
	ret.PriceChargedStr = types.StringPointerValue(data.PriceChargedStr)
	ret.Billing = types.StringPointerValue(data.Billing)
	ret.BillingInfo = types.StringPointerValue(data.BillingInfo)
	ret.Date = types.StringPointerValue(data.Date)
	// ret.DateValue = types.Int64PointerValue(int64(data.DateValue))
	ret.BillingPeriod = types.StringPointerValue(data.BillingPeriod)
	ret.IsFreeTrial = types.StringPointerValue(data.IsFreeTrial)
	ret.Price = types.StringPointerValue(data.Price)
	ret.PriceAndTax = types.Float64PointerValue(data.PriceAndTax)
	ret.Duration = types.StringPointerValue(data.Duration)
	ret.IsFree = types.StringPointerValue(data.IsFree)
	return ret
}
func VoucheringPolicyResourceModelFrom(data piano_publisher.VoucheringPolicy) VoucheringPolicyResourceModel {
	ret := VoucheringPolicyResourceModel{}
	ret.VoucheringPolicyRedemptionUrl = types.StringValue(data.VoucheringPolicyRedemptionUrl)
	ret.VoucheringPolicyBillingPlanDescription = types.StringValue(data.VoucheringPolicyBillingPlanDescription)
	ret.VoucheringPolicyBillingPlan = types.StringValue(data.VoucheringPolicyBillingPlan)
	ret.VoucheringPolicyId = types.StringValue(data.VoucheringPolicyId)
	return ret
}

func TermBriefResourceModelFrom(data piano_publisher.TermBrief) TermBriefResourceModel {
	ret := TermBriefResourceModel{}
	ret.Disabled = types.BoolValue(data.Disabled)
	ret.Name = types.StringValue(data.Name)
	ret.TermId = types.StringValue(data.TermId)
	return ret
}
func ResourceResourceModelFrom(data piano_publisher.Resource) ResourceResourceModel {
	ret := ResourceResourceModel{}
	ret.Disabled = types.BoolValue(data.Disabled)
	ret.ResourceUrl = types.StringPointerValue(data.ResourceUrl)
	ret.PublishDate = types.Int64Value(int64(data.PublishDate))
	ret.ExternalId = types.StringPointerValue(data.ExternalId)
	ret.IsFbiaResource = types.BoolValue(data.IsFbiaResource)
	ret.CreateDate = types.Int64Value(int64(data.CreateDate))
	ret.Name = types.StringValue(data.Name)
	ret.Rid = types.StringValue(data.Rid)
	ret.Deleted = types.BoolValue(data.Deleted)
	ret.Type = types.StringValue(string(data.Type))
	ret.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	ret.Description = types.StringPointerValue(data.Description)
	ret.Aid = types.StringValue(data.Aid)
	ret.PurchaseUrl = types.StringPointerValue(data.PurchaseUrl)
	ret.ImageUrl = types.StringPointerValue(data.ImageUrl)
	ret.BundleType = types.StringPointerValue((*string)(data.BundleType))
	return ret
}

func (r *PaymentTermResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}
func (r *PaymentTermResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *PaymentTermResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PaymentTermResourceModel

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
	state.ProductCategory = types.StringValue(data.ProductCategory)
	state.CurrencySymbol = types.StringValue(data.CurrencySymbol)
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
	state.Resource = &Resource
	state.EvtVerificationPeriod = types.Int32PointerValue(data.EvtVerificationPeriod)
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.VerifyOnRenewal = types.BoolValue(data.VerifyOnRenewal)
	state.BillingConfig = types.StringValue(data.BillingConfig)
	state.PaymentNewCustomersOnly = types.BoolValue(data.PaymentNewCustomersOnly)
	state.TermBillingDescriptor = types.StringValue(data.TermBillingDescriptor)
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	state.CollectAddress = types.BoolValue(data.CollectAddress)
	state.ScheduleBilling = types.StringPointerValue(data.ScheduleBilling)
	state.PaymentHasFreeTrial = types.BoolValue(data.PaymentHasFreeTrial)
	state.Aid = types.StringValue(data.Aid)

	changeOptionsElements := []TermChangeOptionResourceModel{}
	for _, element := range data.ChangeOptions {
		changeOptionsElements = append(changeOptionsElements, TermChangeOptionResourceModelFrom(element))
	}
	state.ChangeOptions = changeOptionsElements
	state.PaymentFirstPrice = types.Float64Value(data.PaymentFirstPrice)
	state.PaymentIsSubscription = types.BoolValue(data.PaymentIsSubscription)
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

func (r *PaymentTermResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PaymentTermResourceModel
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
func (r *PaymentTermResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := TermResourceIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Term resource id", fmt.Sprintf("Unable to parse contract resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), id.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("term_id"), id.TermId)...)
}
