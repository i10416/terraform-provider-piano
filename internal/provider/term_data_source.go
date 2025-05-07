// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &TermDataSource{}
	_ datasource.DataSourceWithConfigure = &TermDataSource{}
)

func NewTermDataSource() datasource.DataSource {
	return &TermDataSource{}
}

// TermDataSource defines the data source implementation.
type TermDataSource struct {
	client *piano_publisher.Client
}

type PeriodDataSourceModel struct {
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
type CountryDataSourceModel struct {
	CountryCode types.String            `tfsdk:"country_code"` // The country code
	CountryId   types.String            `tfsdk:"country_id"`   // The country ID
	CountryName types.String            `tfsdk:"country_name"` // The country name
	Regions     []RegionDataSourceModel `tfsdk:"regions"`
}
type AdvancedOptionsDataSourceModel struct {
	ShowOptions []types.String `tfsdk:"show_options"`
}
type TermBriefDataSourceModel struct {
	Disabled types.Bool   `tfsdk:"disabled"` // Whether the term is disabled
	Name     types.String `tfsdk:"name"`     // The term name
	TermId   types.String `tfsdk:"term_id"`  // The term ID
}
type DeliveryZoneDataSourceModel struct {
	Countries        []CountryDataSourceModel   `tfsdk:"countries"`
	DeliveryZoneId   types.String               `tfsdk:"delivery_zone_id"`   // The delivery zone ID
	DeliveryZoneName types.String               `tfsdk:"delivery_zone_name"` // The delivery zone name
	Terms            []TermBriefDataSourceModel `tfsdk:"terms"`
}
type ExternalAPIFieldDataSourceModel struct {
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
type TermDataSourceModel struct {
	AdviewAccessPeriod                    types.Int32                              `tfsdk:"adview_access_period"`  // The access duration (deprecated)
	AdviewVastUrl                         types.String                             `tfsdk:"adview_vast_url"`       // The VAST URL for adview_access_period (deprecated).
	Aid                                   types.String                             `tfsdk:"aid"`                   // The application ID
	AllowStartInFuture                    types.Bool                               `tfsdk:"allow_start_in_future"` // Allow start in the future
	BillingConfig                         types.String                             `tfsdk:"billing_config"`        // The type of billing config
	BillingConfiguration                  types.String                             `tfsdk:"billing_configuration"` // A JSON value representing a list of the access periods with billing configurations (replaced with "payment_billing_plan(String)")
	ChangeOptions                         []TermChangeOptionDataSourceModel        `tfsdk:"change_options"`
	CollectAddress                        types.Bool                               `tfsdk:"collect_address"`              // Whether to collect an address for this term
	CollectShippingAddress                types.Bool                               `tfsdk:"collect_shipping_address"`     // Whether to collect a shipping address for this gift term
	CreateDate                            types.Int64                              `tfsdk:"create_date"`                  // The creation date
	CurrencySymbol                        types.String                             `tfsdk:"currency_symbol"`              // The currency symbol
	CustomDefaultAccessPeriod             types.Int32                              `tfsdk:"custom_default_access_period"` // The default access period
	CustomRequireUser                     types.Bool                               `tfsdk:"custom_require_user"`          // Whether a valid user is required to complete the term (deprecated)
	DefaultCountry                        *CountryDataSourceModel                  `tfsdk:"default_country"`
	DeliveryZone                          []DeliveryZoneDataSourceModel            `tfsdk:"delivery_zone"`
	Description                           types.String                             `tfsdk:"description"`                  // The description of the term
	EvtCdsProductId                       types.String                             `tfsdk:"evt_cds_product_id"`           // The <a href="https://docs.piano.io/external-service-term/#externalcds">CDS</a> product ID.
	EvtFixedTimeAccessPeriod              types.Int32                              `tfsdk:"evt_fixed_time_access_period"` // The period to grant access for (in days)
	EvtGooglePlayProductId                types.String                             `tfsdk:"evt_google_play_product_id"`   // Google Play's product ID
	EvtGracePeriod                        types.Int32                              `tfsdk:"evt_grace_period"`             // The External API grace period
	EvtItunesBundleId                     types.String                             `tfsdk:"evt_itunes_bundle_id"`         // iTunes's bundle ID
	EvtItunesProductId                    types.String                             `tfsdk:"evt_itunes_product_id"`        // iTunes's product ID
	EvtVerificationPeriod                 types.Int32                              `tfsdk:"evt_verification_period"`      // The <a href = "https://docs.piano.io/external-service-term/#externaltermverification">periodicity</a> (in seconds) of checking the EVT subscription with the external service
	ExternalApiFormFields                 []ExternalAPIFieldDataSourceModel        `tfsdk:"external_api_form_fields"`
	ExternalApiId                         types.String                             `tfsdk:"external_api_id"`                              // The ID of the external API configuration
	ExternalApiName                       types.String                             `tfsdk:"external_api_name"`                            // The name of the external API configuration
	ExternalApiSource                     types.Int32                              `tfsdk:"external_api_source"`                          // The source of the external API configuration
	ExternalProductIds                    types.String                             `tfsdk:"external_product_ids"`                         // <a href="https://docs.piano.io/linked-term/#external-product">“External products"</a> are entities of the external system accessed by users. If you enter multiple values (separated by a comma), Piano will create a standard resource for each product and also a bundled resource that will group them. Example: "digital_prod,print_sub_access,main_articles".
	ExternalTermId                        types.String                             `tfsdk:"external_term_id"`                             // The ID of the term in the external system. Provided by the external system.
	IsAllowedToChangeSchedulePeriodInPast types.Bool                               `tfsdk:"is_allowed_to_change_schedule_period_in_past"` // Whether the term allows to change its schedule period created previously
	MaximumDaysInAdvance                  types.Int32                              `tfsdk:"maximum_days_in_advance"`                      // Maximum days in advance
	Name                                  types.String                             `tfsdk:"name"`                                         // The term name
	PaymentAllowGift                      types.Bool                               `tfsdk:"payment_allow_gift"`                           // Whether the term can be gifted
	PaymentAllowPromoCodes                types.Bool                               `tfsdk:"payment_allow_promo_codes"`                    // Whether to allow promo codes to be applied
	PaymentAllowRenewDays                 types.Int32                              `tfsdk:"payment_allow_renew_days"`                     // How many days in advance users user can renew
	PaymentBillingPlan                    types.String                             `tfsdk:"payment_billing_plan"`                         // The billing plan for the term
	PaymentBillingPlanDescription         types.String                             `tfsdk:"payment_billing_plan_description"`             // The description of the term billing plan
	PaymentBillingPlanTable               []PaymentBillingPlanTableDataSourceModel `tfsdk:"payment_billing_plan_table"`
	PaymentCurrency                       types.String                             `tfsdk:"payment_currency"`                  // The currency of the term
	PaymentFirstPrice                     types.Float64                            `tfsdk:"payment_first_price"`               // The first price of the term
	PaymentForceAutoRenew                 types.Bool                               `tfsdk:"payment_force_auto_renew"`          // Prevents users from disabling autorenewal (always "TRUE" for dynamic terms)
	PaymentHasFreeTrial                   types.Bool                               `tfsdk:"payment_has_free_trial"`            // Whether payment includes a free trial
	PaymentIsCustomPriceAvailable         types.Bool                               `tfsdk:"payment_is_custom_price_available"` // Whether users can pay more than term price
	PaymentIsSubscription                 types.Bool                               `tfsdk:"payment_is_subscription"`           // Whether this term (payment or dynamic) is a subscription (unlike one-off)
	PaymentNewCustomersOnly               types.Bool                               `tfsdk:"payment_new_customers_only"`        // Whether to show the term only to users having no dynamic or purchase conversions yet
	PaymentRenewGracePeriod               types.Int32                              `tfsdk:"payment_renew_grace_period"`        // The number of days after expiration to still allow access to the resource
	PaymentTrialNewCustomersOnly          types.Bool                               `tfsdk:"payment_trial_new_customers_only"`  // Whether to allow trial period only to users having no purchases yet
	ProductCategory                       types.String                             `tfsdk:"product_category"`                  // The product category
	RegistrationAccessPeriod              types.Int32                              `tfsdk:"registration_access_period"`        // The access duration (in seconds) for the registration term
	RegistrationGracePeriod               types.Int32                              `tfsdk:"registration_grace_period"`         // How long (in seconds) after registration users can get access to the term
	Resource                              *ResourceDataSourceModel                 `tfsdk:"resource"`
	Schedule                              *ScheduleDataSourceModel                 `tfsdk:"schedule"`
	ScheduleBilling                       types.String                             `tfsdk:"schedule_billing"`            // The schedule billing
	SharedAccountCount                    types.Int32                              `tfsdk:"shared_account_count"`        // The count of allowed shared-subscription accounts
	SharedRedemptionUrl                   types.String                             `tfsdk:"shared_redemption_url"`       // The shared subscription redemption URL
	ShowFullBillingPlan                   types.Bool                               `tfsdk:"show_full_billing_plan"`      // Show full billing plan on checkout for the dynamic term
	SubscriptionManagementUrl             types.String                             `tfsdk:"subscription_management_url"` // A link to the external system where users can manage their subscriptions (similar to Piano’s MyAccount).
	TermBillingDescriptor                 types.String                             `tfsdk:"term_billing_descriptor"`     // The term billing descriptor
	TermId                                types.String                             `tfsdk:"term_id"`                     // The term ID
	Type                                  types.String                             `tfsdk:"type"`                        // The term type
	TypeName                              types.String                             `tfsdk:"type_name"`                   // The term type name
	UpdateDate                            types.Int64                              `tfsdk:"update_date"`                 // The update date
	VerifyOnRenewal                       types.Bool                               `tfsdk:"verify_on_renewal"`           // Whether the term should be verified before renewal (if "FALSE", this step is skipped)
	VoucheringPolicy                      *VoucheringPolicyDataSourceModel         `tfsdk:"vouchering_policy"`
}
type VoucheringPolicyDataSourceModel struct {
	VoucheringPolicyBillingPlan            types.String `tfsdk:"vouchering_policy_billing_plan"`             // The billing plan of the vouchering policy
	VoucheringPolicyBillingPlanDescription types.String `tfsdk:"vouchering_policy_billing_plan_description"` // The description of the vouchering policy billing plan
	VoucheringPolicyId                     types.String `tfsdk:"vouchering_policy_id"`                       // The vouchering policy ID
	VoucheringPolicyRedemptionUrl          types.String `tfsdk:"vouchering_policy_redemption_url"`           // The vouchering policy redemption URL
}
type LightOfferDataSourceModel struct {
	Name    types.String `tfsdk:"name"`     // The offer name
	OfferId types.String `tfsdk:"offer_id"` // The offer ID
}
type TermChangeOptionDataSourceModel struct {
	AdvancedOptions    *AdvancedOptionsDataSourceModel `tfsdk:"advanced_options"`
	BillingTiming      types.String                    `tfsdk:"billing_timing"`        // The billing timing(0: immediate term change;1: term change at the end of the current cycle;2: term change on the next sell date;3: term change at the end of the current period)
	CollectAddress     types.Bool                      `tfsdk:"collect_address"`       // Whether to collect an address for this term
	Description        types.String                    `tfsdk:"description"`           // A description of the term change option; provided by the client
	FromBillingPlan    types.String                    `tfsdk:"from_billing_plan"`     // The "From" billing plan
	FromPeriodId       types.String                    `tfsdk:"from_period_id"`        // The ID of the "From" term period
	FromPeriodName     types.String                    `tfsdk:"from_period_name"`      // The name of the "From" term period
	FromResourceId     types.String                    `tfsdk:"from_resource_id"`      // The ID of the "From" resource
	FromResourceName   types.String                    `tfsdk:"from_resource_name"`    // The name of the "From" resource
	FromScheduled      types.Bool                      `tfsdk:"from_scheduled"`        // Whether the subscription is upgraded from a scheduled term
	FromTermId         types.String                    `tfsdk:"from_term_id"`          // The ID of the "From" term
	FromTermName       types.String                    `tfsdk:"from_term_name"`        // The name of the "From" term
	ImmediateAccess    types.Bool                      `tfsdk:"immediate_access"`      // Whether the access begins immediately
	IncludeTrial       types.Bool                      `tfsdk:"include_trial"`         // Whether trial is enabled (not in use, always "FALSE")
	ProrateAccess      types.Bool                      `tfsdk:"prorate_access"`        // Whether the <a href="https://docs.piano.io/upgrades/?paragraphId=b27954ef84407e4#prorate-billing-amount">Prorate billing amount</a> function is enabled
	SharedAccountCount types.Int32                     `tfsdk:"shared_account_count"`  // The count of allowed shared-subscription accounts
	TermChangeOptionId types.String                    `tfsdk:"term_change_option_id"` // The ID of the term change option
	ToBillingPlan      types.String                    `tfsdk:"to_billing_plan"`       // The "To" billing plan
	ToPeriodId         types.String                    `tfsdk:"to_period_id"`          // The ID of the "To" term period
	ToPeriodName       types.String                    `tfsdk:"to_period_name"`        // The period name of the "To" term
	ToResourceId       types.String                    `tfsdk:"to_resource_id"`        // The ID of the "To" resource
	ToResourceName     types.String                    `tfsdk:"to_resource_name"`      // The name of the "To" resource
	ToScheduled        types.Bool                      `tfsdk:"to_scheduled"`          // Whether the subscription is upgraded to a scheduled term
	ToTermId           types.String                    `tfsdk:"to_term_id"`            // The ID of the "To" term
	ToTermName         types.String                    `tfsdk:"to_term_name"`          // The name of the "To" term
	UpgradeOffers      []LightOfferDataSourceModel     `tfsdk:"upgrade_offers"`
}

type ScheduleDataSourceModel struct {
	Aid        types.String            `tfsdk:"aid"`         // The application ID
	CreateDate types.Int64             `tfsdk:"create_date"` // The creation date
	Deleted    types.Bool              `tfsdk:"deleted"`     // Whether the object is deleted
	Name       types.String            `tfsdk:"name"`        // The schedule name
	Periods    []PeriodDataSourceModel `tfsdk:"periods"`
	ScheduleId types.String            `tfsdk:"schedule_id"` // The schedule ID
	UpdateDate types.Int64             `tfsdk:"update_date"` // The update date
}
type PaymentBillingPlanTableDataSourceModel struct {
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
type RegionDataSourceModel struct {
	RegionCode types.String `tfsdk:"region_code"` // The code of the country region
	RegionId   types.String `tfsdk:"region_id"`   // The ID of the country region
	RegionName types.String `tfsdk:"region_name"` // The name of the country region
}

func (r *TermDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_term"
}
func (*TermDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Term datasource. Term is a pair of price and schedule that is a part of offers.",
		Attributes: map[string]schema.Attribute{
			"payment_allow_renew_days": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "How many days in advance users user can renew",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The description of the term",
			},
			"payment_allow_promo_codes": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to allow promo codes to be applied",
			},
			"payment_billing_plan_description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The description of the term billing plan",
			},
			"maximum_days_in_advance": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum days in advance",
			},
			"delivery_zone": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"delivery_zone_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The delivery zone ID",
						},
						"delivery_zone_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The delivery zone name",
						},
						"countries": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"country_name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The country name",
									},
									"country_code": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The country code",
									},
									"country_id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The country ID",
									},
									"regions": schema.ListNestedAttribute{
										Computed: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"region_name": schema.StringAttribute{
													Computed:            true,
													MarkdownDescription: "The name of the country region",
												},
												"region_code": schema.StringAttribute{
													Computed:            true,
													MarkdownDescription: "The code of the country region",
												},
												"region_id": schema.StringAttribute{
													Computed:            true,
													MarkdownDescription: "The ID of the country region",
												},
											},
										},
									},
								},
							},
						},
						"terms": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"term_id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The term ID",
									},
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The term name",
									},
									"disabled": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "Whether the term is disabled",
									},
								},
							},
						},
					},
				},
			},
			"adview_access_period": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The access duration (deprecated)",
			},
			"is_allowed_to_change_schedule_period_in_past": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the term allows to change its schedule period created previously",
			},
			"payment_is_custom_price_available": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether users can pay more than term price",
			},
			"term_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The term ID",
			},
			"vouchering_policy": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"vouchering_policy_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The vouchering policy ID",
					},
					"vouchering_policy_billing_plan": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The billing plan of the vouchering policy",
					},
					"vouchering_policy_billing_plan_description": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The description of the vouchering policy billing plan",
					},
					"vouchering_policy_redemption_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The vouchering policy redemption URL",
					},
				},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The term name",
			},
			"evt_itunes_product_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "iTunes's product ID",
			},
			"adview_vast_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The VAST URL for adview_access_period (deprecated).",
			},
			"subscription_management_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A link to the external system where users can manage their subscriptions (similar to Piano’s MyAccount).",
			},
			"billing_configuration": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A JSON value representing a list of the access periods with billing configurations (replaced with \"payment_billing_plan(String)\")",
			},
			"payment_is_subscription": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether this term (payment or dynamic) is a subscription (unlike one-off)",
			},
			"custom_default_access_period": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The default access period",
			},
			"payment_first_price": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "The first price of the term",
			},
			"registration_access_period": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The access duration (in seconds) for the registration term",
			},
			"change_options": schema.ListNestedAttribute{
				Computed: true,
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
								"show_options": schema.SetAttribute{
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
						"upgrade_offers": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"offer_id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The offer ID",
									},
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The offer name",
									},
								},
							},
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
							Computed:            true,
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
			"payment_has_free_trial": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether payment includes a free trial",
			},
			"schedule_billing": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The schedule billing",
			},
			"collect_address": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to collect an address for this term",
			},
			"external_api_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the external API configuration",
			},
			"update_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The update date",
			},
			"allow_start_in_future": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Allow start in the future",
			},
			"registration_grace_period": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "How long (in seconds) after registration users can get access to the term",
			},
			"external_product_ids": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "<a href=\"https://docs.piano.io/linked-term/#external-product\">“External products\"</a> are entities of the external system accessed by users. If you enter multiple values (separated by a comma), Piano will create a standard resource for each product and also a bundled resource that will group them. Example: \"digital_prod,print_sub_access,main_articles\".",
			},
			"evt_cds_product_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The <a href=\"https://docs.piano.io/external-service-term/#externalcds\">CDS</a> product ID.",
			},
			"collect_shipping_address": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to collect a shipping address for this gift term",
			},
			"term_billing_descriptor": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The term billing descriptor",
			},
			"payment_new_customers_only": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to show the term only to users having no dynamic or purchase conversions yet",
			},
			"billing_config": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of billing config",
			},
			"verify_on_renewal": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the term should be verified before renewal (if \"FALSE\", this step is skipped)",
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
			"payment_billing_plan_table": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"is_free": schema.StringAttribute{
							Computed: true,
						},
						"duration": schema.StringAttribute{
							Computed: true,
						},
						"price_and_tax": schema.Float64Attribute{
							Computed: true,
						},
						"price": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "price with currency unit symbol",
						},
						"is_free_trial": schema.StringAttribute{
							Computed: true,
						},
						"billing_period": schema.StringAttribute{
							Computed: true,
						},
						"date_value": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Payment billing plan table date in timestamp",
						},
						"date": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Payment billing plan table date for humans such as \"Today\" or \"Apr 17, 2026\"",
						},
						"billing_info": schema.StringAttribute{
							Computed: true,
						},
						"billing": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "payment condition such as \"one payment of $99.99\" or \"$119.99 per year\"",
						},
						"price_charged_str": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "price with currency unit symbol",
						},
						"period": schema.StringAttribute{
							Computed: true,
						},
						"currency": schema.StringAttribute{
							Computed: true,
						},
						"total_billing": schema.StringAttribute{
							Computed: true,
						},
						"is_pay_what_you_want": schema.StringAttribute{
							Computed: true,
						},
						"short_period": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "human readable billing period in shorter expression such as /yr",
						},
						"price_and_tax_in_minor_unit": schema.Float32Attribute{
							Computed: true,
						},
						"price_value": schema.Float64Attribute{
							Computed: true,
						},
						"cycles": schema.StringAttribute{
							Computed: true,
						},
						"is_trial": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"custom_require_user": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether a valid user is required to complete the term (deprecated)",
			},
			"payment_billing_plan": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The billing plan for the term",
			},
			"payment_allow_gift": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the term can be gifted",
			},
			"payment_force_auto_renew": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Prevents users from disabling autorenewal (always \"TRUE\" for dynamic terms)",
			},
			"schedule": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The schedule name",
					},
					"update_date": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The update date",
					},
					"create_date": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The creation date",
					},
					"schedule_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The schedule ID",
					},
					"deleted": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether the object is deleted",
					},
					"aid": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The application ID",
					},
					"periods": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"is_sale_started": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether sale is started for the period",
								},
								"period_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The period ID",
								},
								"sell_date": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The sell date of the period",
								},
								"update_date": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The update date",
								},
								"create_date": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The creation date",
								},
								"end_date": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The date when the period ends",
								},
								"begin_date": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The date when the period begins",
								},
								"deleted": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether the object is deleted",
								},
								"name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The period name",
								},
								"is_active": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether the period is active. A period is in the Active state when the sell date is passed but the end date is not reached",
								},
							},
						},
					},
				},
			},
			"payment_trial_new_customers_only": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to allow trial period only to users having no purchases yet",
			},
			"payment_currency": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The currency of the term",
			},
			"shared_redemption_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The shared subscription redemption URL",
			},
			"currency_symbol": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The currency symbol",
			},
			"type_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The term type name",
				Validators: []validator.String{
					stringvalidator.OneOf("Payment", "Ad View", "Registration", "Newsletter", "External", "Custom", "Access Granted", "Gift", "Specific Email Addresses Contract", "Email Domain Contract", "IP Range Contract", "Dynamic", "Linked"),
				},
			},
			"product_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The product category",
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
			"show_full_billing_plan": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Show full billing plan on checkout for the dynamic term",
			},
			"default_country": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"country_name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The country name",
					},
					"country_code": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The country code",
					},
					"country_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The country ID",
					},
					"regions": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"region_name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The name of the country region",
								},
								"region_code": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The code of the country region",
								},
								"region_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The ID of the country region",
								},
							},
						},
					},
				},
			},
			"payment_renew_grace_period": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The number of days after expiration to still allow access to the resource",
			},
			"evt_itunes_bundle_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "iTunes's bundle ID",
			},
			"external_term_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the term in the external system. Provided by the external system.",
			},
		},
	}
}
func ExternalAPIFieldDataSourceModelFrom(data piano_publisher.ExternalAPIField) ExternalAPIFieldDataSourceModel {
	ret := ExternalAPIFieldDataSourceModel{}
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
func RegionDataSourceModelFrom(data piano_publisher.Region) RegionDataSourceModel {
	ret := RegionDataSourceModel{}
	ret.RegionId = types.StringValue(data.RegionId)
	ret.RegionCode = types.StringValue(data.RegionCode)
	ret.RegionName = types.StringValue(data.RegionName)
	return ret
}
func PeriodDataSourceModelFrom(data piano_publisher.Period) PeriodDataSourceModel {
	ret := PeriodDataSourceModel{}
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
func LightOfferDataSourceModelFrom(data piano_publisher.LightOffer) LightOfferDataSourceModel {
	ret := LightOfferDataSourceModel{}
	ret.Name = types.StringValue(data.Name)
	ret.OfferId = types.StringValue(data.OfferId)
	return ret
}
func AdvancedOptionsDataSourceModelFrom(data piano_publisher.AdvancedOptions) AdvancedOptionsDataSourceModel {
	ret := AdvancedOptionsDataSourceModel{}
	showOptionsElements := []types.String{}
	for _, element := range data.ShowOptions {
		showOptionsElements = append(showOptionsElements, types.StringValue(element))
	}
	sort.Slice(showOptionsElements, func(i, j int) bool {
		return showOptionsElements[i].ValueString() < showOptionsElements[j].ValueString()
	})
	ret.ShowOptions = showOptionsElements
	return ret
}
func TermChangeOptionDataSourceModelFrom(data piano_publisher.TermChangeOption) TermChangeOptionDataSourceModel {
	ret := TermChangeOptionDataSourceModel{}
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
	upgradeOffersElements := []LightOfferDataSourceModel{}
	for _, element := range data.UpgradeOffers {
		upgradeOffersElements = append(upgradeOffersElements, LightOfferDataSourceModelFrom(element))
	}
	ret.UpgradeOffers = upgradeOffersElements
	ret.FromBillingPlan = types.StringValue(data.FromBillingPlan)
	ret.FromTermId = types.StringValue(data.FromTermId)
	ret.ImmediateAccess = types.BoolValue(data.ImmediateAccess)
	ret.IncludeTrial = types.BoolValue(data.IncludeTrial)
	ret.ToTermId = types.StringValue(data.ToTermId)
	ret.ToTermName = types.StringValue(data.ToTermName)
	ret.TermChangeOptionId = types.StringValue(data.TermChangeOptionId)
	AdvancedOptions := AdvancedOptionsDataSourceModelFrom(data.AdvancedOptions)
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
func ScheduleDataSourceModelFrom(data piano_publisher.Schedule) ScheduleDataSourceModel {
	ret := ScheduleDataSourceModel{}
	periodsElements := []PeriodDataSourceModel{}
	for _, element := range data.Periods {
		periodsElements = append(periodsElements, PeriodDataSourceModelFrom(element))
	}
	ret.Periods = periodsElements
	ret.Aid = types.StringValue(data.Aid)
	ret.Deleted = types.BoolValue(data.Deleted)
	ret.ScheduleId = types.StringValue(data.ScheduleId)
	ret.CreateDate = types.Int64Value(int64(data.CreateDate))
	ret.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	ret.Name = types.StringValue(data.Name)
	return ret
}
func DeliveryZoneDataSourceModelFrom(data piano_publisher.DeliveryZone) DeliveryZoneDataSourceModel {
	ret := DeliveryZoneDataSourceModel{}
	termsElements := []TermBriefDataSourceModel{}
	for _, element := range data.Terms {
		termsElements = append(termsElements, TermBriefDataSourceModelFrom(element))
	}
	ret.Terms = termsElements
	countriesElements := []CountryDataSourceModel{}
	for _, element := range data.Countries {
		countriesElements = append(countriesElements, CountryDataSourceModelFrom(element))
	}
	ret.Countries = countriesElements
	ret.DeliveryZoneName = types.StringValue(data.DeliveryZoneName)
	ret.DeliveryZoneId = types.StringValue(data.DeliveryZoneId)
	return ret
}
func PaymentBillingPlanTableDataSourceModelFrom(data piano_publisher.PaymentBillingPlanTable) PaymentBillingPlanTableDataSourceModel {
	ret := PaymentBillingPlanTableDataSourceModel{}
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
func VoucheringPolicyDataSourceModelFrom(data piano_publisher.VoucheringPolicy) VoucheringPolicyDataSourceModel {
	ret := VoucheringPolicyDataSourceModel{}
	ret.VoucheringPolicyRedemptionUrl = types.StringValue(data.VoucheringPolicyRedemptionUrl)
	ret.VoucheringPolicyBillingPlanDescription = types.StringValue(data.VoucheringPolicyBillingPlanDescription)
	ret.VoucheringPolicyBillingPlan = types.StringValue(data.VoucheringPolicyBillingPlan)
	ret.VoucheringPolicyId = types.StringValue(data.VoucheringPolicyId)
	return ret
}
func CountryDataSourceModelFrom(data piano_publisher.Country) CountryDataSourceModel {
	ret := CountryDataSourceModel{}
	regionsElements := []RegionDataSourceModel{}
	for _, element := range data.Regions {
		regionsElements = append(regionsElements, RegionDataSourceModelFrom(element))
	}
	ret.Regions = regionsElements
	ret.CountryId = types.StringValue(data.CountryId)
	ret.CountryCode = types.StringValue(data.CountryCode)
	ret.CountryName = types.StringValue(data.CountryName)
	return ret
}
func TermBriefDataSourceModelFrom(data piano_publisher.TermBrief) TermBriefDataSourceModel {
	ret := TermBriefDataSourceModel{}
	ret.Disabled = types.BoolValue(data.Disabled)
	ret.Name = types.StringValue(data.Name)
	ret.TermId = types.StringValue(data.TermId)
	return ret
}
func ResourceDataSourceModelFrom(data piano_publisher.Resource) ResourceDataSourceModel {
	ret := ResourceDataSourceModel{}
	ret.Disabled = types.BoolValue(data.Disabled)
	ret.ResourceUrl = types.StringPointerValue(data.ResourceUrl)
	ret.PublishDate = types.Int64Value(int64(data.PublishDate))
	ret.ExternalId = types.StringPointerValue(data.ExternalId)
	ret.TypeLabel = types.StringValue(string(data.TypeLabel))
	ret.IsFbiaResource = types.BoolValue(data.IsFbiaResource)
	ret.CreateDate = types.Int64Value(int64(data.CreateDate))
	ret.Name = types.StringValue(data.Name)
	ret.Rid = types.StringValue(data.Rid)
	ret.Deleted = types.BoolValue(data.Deleted)
	ret.Type = types.StringValue(string(data.Type))
	ret.BundleTypeLabel = types.StringPointerValue((*string)(data.BundleTypeLabel))
	ret.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	ret.Description = types.StringPointerValue(data.Description)
	ret.Aid = types.StringValue(data.Aid)
	ret.PurchaseUrl = types.StringPointerValue(data.PurchaseUrl)
	ret.ImageUrl = types.StringPointerValue(data.ImageUrl)
	ret.BundleType = types.StringPointerValue((*string)(data.BundleType))
	return ret
}

func (d *TermDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*piano_publisher.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *piano.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *TermDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TermDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := d.client.GetPublisherTermGet(ctx, &piano_publisher.GetPublisherTermGetParams{
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

	state.ExternalTermId = types.StringPointerValue(data.ExternalTermId)
	state.EvtItunesBundleId = types.StringPointerValue(data.EvtItunesBundleId)
	state.PaymentRenewGracePeriod = types.Int32Value(data.PaymentRenewGracePeriod)
	if data.DefaultCountry != nil {
		DefaultCountry := CountryDataSourceModelFrom(*data.DefaultCountry)
		state.DefaultCountry = &DefaultCountry
	}
	state.ShowFullBillingPlan = types.BoolPointerValue(data.ShowFullBillingPlan)
	state.ExternalApiId = types.StringPointerValue(data.ExternalApiId)
	state.Type = types.StringValue(string(data.Type))
	state.ProductCategory = types.StringValue(data.ProductCategory)
	state.TypeName = types.StringValue(string(data.TypeName))
	state.CurrencySymbol = types.StringValue(data.CurrencySymbol)
	state.SharedRedemptionUrl = types.StringPointerValue(data.SharedRedemptionUrl)
	state.PaymentCurrency = types.StringValue(data.PaymentCurrency)
	state.PaymentTrialNewCustomersOnly = types.BoolValue(data.PaymentTrialNewCustomersOnly)
	if data.Schedule != nil {
		Schedule := ScheduleDataSourceModelFrom(*data.Schedule)
		state.Schedule = &Schedule
	}
	state.PaymentForceAutoRenew = types.BoolValue(data.PaymentForceAutoRenew)
	state.PaymentAllowGift = types.BoolValue(data.PaymentAllowGift)
	state.PaymentBillingPlan = types.StringValue(data.PaymentBillingPlan)
	state.CustomRequireUser = types.BoolPointerValue(data.CustomRequireUser)
	paymentBillingPlanTableElements := []PaymentBillingPlanTableDataSourceModel{}
	for _, element := range data.PaymentBillingPlanTable {
		paymentBillingPlanTableElements = append(paymentBillingPlanTableElements, PaymentBillingPlanTableDataSourceModelFrom(element))
	}
	state.PaymentBillingPlanTable = paymentBillingPlanTableElements
	state.EvtGracePeriod = types.Int32PointerValue(data.EvtGracePeriod)
	state.EvtFixedTimeAccessPeriod = types.Int32PointerValue(data.EvtFixedTimeAccessPeriod)
	Resource := ResourceDataSourceModelFrom(data.Resource)
	state.Resource = &Resource
	state.EvtGooglePlayProductId = types.StringPointerValue(data.EvtGooglePlayProductId)
	state.EvtVerificationPeriod = types.Int32PointerValue(data.EvtVerificationPeriod)
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.VerifyOnRenewal = types.BoolValue(data.VerifyOnRenewal)
	state.BillingConfig = types.StringValue(data.BillingConfig)
	state.PaymentNewCustomersOnly = types.BoolValue(data.PaymentNewCustomersOnly)
	state.TermBillingDescriptor = types.StringValue(data.TermBillingDescriptor)
	state.CollectShippingAddress = types.BoolPointerValue(data.CollectShippingAddress)
	state.EvtCdsProductId = types.StringPointerValue(data.EvtCdsProductId)
	state.ExternalProductIds = types.StringPointerValue(data.ExternalProductIds)
	state.RegistrationGracePeriod = types.Int32PointerValue(data.RegistrationGracePeriod)
	state.AllowStartInFuture = types.BoolPointerValue(data.AllowStartInFuture)
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	state.ExternalApiName = types.StringPointerValue(data.ExternalApiName)
	state.CollectAddress = types.BoolValue(data.CollectAddress)
	state.ScheduleBilling = types.StringPointerValue(data.ScheduleBilling)
	state.PaymentHasFreeTrial = types.BoolValue(data.PaymentHasFreeTrial)
	state.ExternalApiSource = types.Int32PointerValue((*int32)(data.ExternalApiSource))
	state.Aid = types.StringValue(data.Aid)
	if data.ExternalApiFormFields != nil {
		externalApiFormFieldsElements := []ExternalAPIFieldDataSourceModel{}
		for _, element := range *data.ExternalApiFormFields {
			externalApiFormFieldsElements = append(externalApiFormFieldsElements, ExternalAPIFieldDataSourceModelFrom(element))
		}
		state.ExternalApiFormFields = externalApiFormFieldsElements
	}
	state.SharedAccountCount = types.Int32PointerValue(data.SharedAccountCount)
	changeOptionsElements := []TermChangeOptionDataSourceModel{}
	for _, element := range data.ChangeOptions {
		changeOptionsElements = append(changeOptionsElements, TermChangeOptionDataSourceModelFrom(element))
	}
	state.ChangeOptions = changeOptionsElements
	state.RegistrationAccessPeriod = types.Int32PointerValue(data.RegistrationAccessPeriod)
	state.PaymentFirstPrice = types.Float64Value(data.PaymentFirstPrice)
	state.CustomDefaultAccessPeriod = types.Int32PointerValue(data.CustomDefaultAccessPeriod)
	state.PaymentIsSubscription = types.BoolValue(data.PaymentIsSubscription)
	state.BillingConfiguration = types.StringPointerValue(data.BillingConfiguration)
	state.SubscriptionManagementUrl = types.StringValue(data.SubscriptionManagementUrl)
	state.AdviewVastUrl = types.StringPointerValue(data.AdviewVastUrl)
	state.EvtItunesProductId = types.StringPointerValue(data.EvtItunesProductId)
	state.Name = types.StringValue(data.Name)
	if data.VoucheringPolicy != nil {
		VoucheringPolicy := VoucheringPolicyDataSourceModelFrom(*data.VoucheringPolicy)
		state.VoucheringPolicy = &VoucheringPolicy
	}
	state.TermId = types.StringValue(data.TermId)
	state.PaymentIsCustomPriceAvailable = types.BoolValue(data.PaymentIsCustomPriceAvailable)
	state.IsAllowedToChangeSchedulePeriodInPast = types.BoolValue(data.IsAllowedToChangeSchedulePeriodInPast)
	state.AdviewAccessPeriod = types.Int32PointerValue(data.AdviewAccessPeriod)
	if data.DeliveryZone != nil {
		deliveryZoneElements := []DeliveryZoneDataSourceModel{}
		for _, element := range *data.DeliveryZone {
			deliveryZoneElements = append(deliveryZoneElements, DeliveryZoneDataSourceModelFrom(element))
		}
		state.DeliveryZone = deliveryZoneElements
	}
	state.MaximumDaysInAdvance = types.Int32PointerValue(data.MaximumDaysInAdvance)
	state.PaymentBillingPlanDescription = types.StringValue(data.PaymentBillingPlanDescription)
	state.PaymentAllowPromoCodes = types.BoolValue(data.PaymentAllowPromoCodes)
	state.Description = types.StringValue(data.Description)
	state.PaymentAllowRenewDays = types.Int32Value(data.PaymentAllowRenewDays)

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
