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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ datasource.DataSource              = &PromotionDataSource{}
	_ datasource.DataSourceWithConfigure = &PromotionDataSource{}
)

// PromotionDataSource defines the resource implementation.
type PromotionDataSource struct {
	client *piano_publisher.Client
}

func NewPromotionDataSource() datasource.DataSource {
	return &PromotionDataSource{}
}
func (r *PromotionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*piano_publisher.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *piano_publisher.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}
func (r *PromotionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_promotion"
}

func PromotionFixedDiscountAttrType() attr.Type {
	return basetypes.ObjectType{
		AttrTypes: map[string]attr.Type{
			"fixed_discount_id": types.StringType,
			"currency":          types.StringType,
			"amount":            types.StringType,
			"amount_value":      types.Float64Type,
		},
	}
}
func PromotionAttrType() attr.Type {
	return basetypes.ObjectType{
		AttrTypes: map[string]attr.Type{
			"discount_type":                types.StringType,
			"name":                         types.StringType,
			"start_date":                   types.Int64Type,
			"term_dependency_type":         types.StringType,
			"create_date":                  types.Int64Type,
			"deleted":                      types.BoolType,
			"update_by":                    types.StringType,
			"aid":                          types.StringType,
			"fixed_promotion_code":         types.StringType,
			"uses":                         types.Int32Type,
			"billing_period_limit":         types.Int32Type,
			"can_be_applied_on_renewal":    types.BoolType,
			"discount_currency":            types.StringType,
			"update_date":                  types.Int64Type,
			"apply_to_all_billing_periods": types.BoolType,
			"never_allow_zero":             types.BoolType,
			"end_date":                     types.Int64Type,
			"fixed_discount_list": types.ListType{
				ElemType: PromotionFixedDiscountAttrType(),
			},
			"new_customers_only":    types.BoolType,
			"status":                types.StringType,
			"percentage_discount":   types.Float64Type,
			"unlimited_uses":        types.BoolType,
			"discount_amount":       types.Float64Type,
			"promotion_id":          types.StringType,
			"promotion_code_prefix": types.StringType,
			"create_by":             types.StringType,
			"uses_allowed":          types.Int32Type,
			"discount":              types.StringType,
		},
	}
}

type PromotionDataSourceModel struct {
	DiscountType             types.String                            `tfsdk:"discount_type"`                // The promotion discount type
	Name                     types.String                            `tfsdk:"name"`                         // The promotion name
	StartDate                types.Int64                             `tfsdk:"start_date"`                   // The start date.
	TermDependencyType       types.String                            `tfsdk:"term_dependency_type"`         // The type of dependency to terms
	CreateDate               types.Int64                             `tfsdk:"create_date"`                  // The creation date
	Deleted                  types.Bool                              `tfsdk:"deleted"`                      // Whether the object is deleted
	UpdateBy                 types.String                            `tfsdk:"update_by"`                    // The last user to update the object
	Aid                      types.String                            `tfsdk:"aid"`                          // The application ID
	FixedPromotionCode       types.String                            `tfsdk:"fixed_promotion_code"`         // The fixed value for all the promotion codes
	Uses                     types.Int32                             `tfsdk:"uses"`                         // How many times the promotion has been used
	BillingPeriodLimit       types.Int32                             `tfsdk:"billing_period_limit"`         // Promotion discount applies to number of billing periods
	CanBeAppliedOnRenewal    types.Bool                              `tfsdk:"can_be_applied_on_renewal"`    // Whether the promotion can be applied on renewal
	DiscountCurrency         types.String                            `tfsdk:"discount_currency"`            // The promotion discount currency
	UpdateDate               types.Int64                             `tfsdk:"update_date"`                  // The update date
	ApplyToAllBillingPeriods types.Bool                              `tfsdk:"apply_to_all_billing_periods"` // Whether to apply the promotion discount to all billing periods ("TRUE")or the first billing period only ("FALSE")
	NeverAllowZero           types.Bool                              `tfsdk:"never_allow_zero"`             // Never allow the value of checkout to be zero
	EndDate                  types.Int64                             `tfsdk:"end_date"`                     // The end date
	FixedDiscountList        []PromotionFixedDiscountDataSourceModel `tfsdk:"fixed_discount_list"`
	NewCustomersOnly         types.Bool                              `tfsdk:"new_customers_only"`    // Whether the promotion allows new customers only
	Status                   types.String                            `tfsdk:"status"`                // The promotion status
	PercentageDiscount       types.Float64                           `tfsdk:"percentage_discount"`   // The promotion discount, percentage
	UnlimitedUses            types.Bool                              `tfsdk:"unlimited_uses"`        // Whether to allow unlimited uses
	DiscountAmount           types.Float64                           `tfsdk:"discount_amount"`       // The promotion discount
	PromotionId              types.String                            `tfsdk:"promotion_id"`          // The promotion ID
	PromotionCodePrefix      types.String                            `tfsdk:"promotion_code_prefix"` // The prefix for all the codes
	CreateBy                 types.String                            `tfsdk:"create_by"`             // The user who created the object
	UsesAllowed              types.Int32                             `tfsdk:"uses_allowed"`          // The number of uses allowed by the promotion
	Discount                 types.String                            `tfsdk:"discount"`              // The promotion discount, formatted
}
type PromotionFixedDiscountDataSourceModel struct {
	FixedDiscountId types.String  `tfsdk:"fixed_discount_id"` // The fixed discount ID
	Currency        types.String  `tfsdk:"currency"`          // The currency of the fixed discount
	Amount          types.String  `tfsdk:"amount"`            // The fixed discount amount
	AmountValue     types.Float64 `tfsdk:"amount_value"`      // The fixed discount amount value
}

func (*PromotionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"discount_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The promotion discount type",
				Validators:          []validator.String{stringvalidator.OneOf("fixed", "percentage")},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The promotion name",
			},
			"start_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The start date.",
			},
			"term_dependency_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of dependency to terms",
				Validators:          []validator.String{stringvalidator.OneOf("all", "include", "unlocked")},
			},
			"create_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The creation date",
			},
			"deleted": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the object is deleted",
			},
			"update_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The last user to update the object",
			},
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			"fixed_promotion_code": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The fixed value for all the promotion codes",
			},
			"uses": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "How many times the promotion has been used",
			},
			"billing_period_limit": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "Promotion discount applies to number of billing periods",
			},
			"can_be_applied_on_renewal": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the promotion can be applied on renewal",
			},
			"discount_currency": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The promotion discount currency",
			},
			"update_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The update date",
			},
			"apply_to_all_billing_periods": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to apply the promotion discount to all billing periods (\"TRUE\")or the first billing period only (\"FALSE\")",
			},
			"never_allow_zero": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Never allow the value of checkout to be zero",
			},
			"end_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The end date",
			},
			"fixed_discount_list": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"fixed_discount_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The fixed discount ID",
						},
						"currency": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The currency of the fixed discount",
						},
						"amount": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The fixed discount amount",
						},
						"amount_value": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "The fixed discount amount value",
						},
					},
				},
			},
			"new_customers_only": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the promotion allows new customers only",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The promotion status",
				Validators:          []validator.String{stringvalidator.OneOf("active", "expired", "new")},
			},
			"percentage_discount": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "The promotion discount, percentage",
			},
			"unlimited_uses": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to allow unlimited uses",
			},
			"discount_amount": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "The promotion discount",
			},
			"promotion_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The promotion ID",
			},
			"promotion_code_prefix": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The prefix for all the codes",
			},
			"create_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user who created the object",
			},
			"uses_allowed": schema.Int32Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The number of uses allowed by the promotion",
			},
			"discount": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The promotion discount, formatted",
			},
		},
	}
}
func PromotionFixedDiscountDataSourceModelFrom(data piano_publisher.PromotionFixedDiscount) PromotionFixedDiscountDataSourceModel {
	ret := PromotionFixedDiscountDataSourceModel{}
	ret.AmountValue = types.Float64Value(data.AmountValue)
	ret.Amount = types.StringValue(data.Amount)
	ret.Currency = types.StringValue(data.Currency)
	ret.FixedDiscountId = types.StringValue(data.FixedDiscountId)
	return ret
}
func (r *PromotionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PromotionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.GetPublisherPromotionGet(ctx, &piano_publisher.GetPublisherPromotionGetParams{
		Aid:         state.Aid.ValueString(),
		PromotionId: state.PromotionId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch promotion, got error: %s", err))
		return
	}
	anyResponse, err := syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.PromotionResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	data := result.Promotion
	state.Discount = types.StringValue(data.Discount)
	state.UsesAllowed = types.Int32PointerValue(data.UsesAllowed)
	state.CreateBy = types.StringValue(data.CreateBy)
	state.PromotionCodePrefix = types.StringPointerValue(data.PromotionCodePrefix)
	state.PromotionId = types.StringValue(data.PromotionId)
	state.DiscountAmount = types.Float64Value(data.DiscountAmount)
	state.UnlimitedUses = types.BoolValue(data.UnlimitedUses)
	state.PercentageDiscount = types.Float64Value(data.PercentageDiscount)
	state.Status = types.StringValue(string(data.Status))
	state.NewCustomersOnly = types.BoolValue(data.NewCustomersOnly)
	fixedDiscountListElements := []PromotionFixedDiscountDataSourceModel{}
	for _, element := range data.FixedDiscountList {
		fixedDiscountListElements = append(fixedDiscountListElements, PromotionFixedDiscountDataSourceModelFrom(element))
	}
	state.FixedDiscountList = fixedDiscountListElements
	state.EndDate = types.Int64Value(int64(data.EndDate))
	state.NeverAllowZero = types.BoolValue(data.NeverAllowZero)
	state.ApplyToAllBillingPeriods = types.BoolValue(data.ApplyToAllBillingPeriods)
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	state.DiscountCurrency = types.StringValue(data.DiscountCurrency)
	state.CanBeAppliedOnRenewal = types.BoolValue(data.CanBeAppliedOnRenewal)
	state.BillingPeriodLimit = types.Int32Value(data.BillingPeriodLimit)
	state.Uses = types.Int32Value(data.Uses)
	state.FixedPromotionCode = types.StringPointerValue(data.FixedPromotionCode)
	state.Aid = types.StringValue(data.Aid)
	state.UpdateBy = types.StringValue(data.UpdateBy)
	state.Deleted = types.BoolValue(data.Deleted)
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.TermDependencyType = types.StringValue(string(data.TermDependencyType))
	state.StartDate = types.Int64Value(int64(data.StartDate))
	state.Name = types.StringValue(data.Name)
	state.DiscountType = types.StringValue(string(data.DiscountType))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
