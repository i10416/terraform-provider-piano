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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &PromotionResource{}
	_ resource.ResourceWithImportState = &PromotionResource{}
)

func NewPromotionResource() resource.Resource {
	return &PromotionResource{}
}

// PromotionResource defines the resource implementation.
type PromotionResource struct {
	client *piano_publisher.Client
}

func (r *PromotionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *PromotionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_promotion"
}

type PromotionResourceModel struct {
	Aid                      types.String                          `tfsdk:"aid"`                          // The application ID
	PromotionId              types.String                          `tfsdk:"promotion_id"`                 // The promotion ID
	Name                     types.String                          `tfsdk:"name"`                         // The promotion name
	StartDate                types.Int64                           `tfsdk:"start_date"`                   // The start date.
	EndDate                  types.Int64                           `tfsdk:"end_date"`                     // The end date
	NewCustomersOnly         types.Bool                            `tfsdk:"new_customers_only"`           // Whether the promotion allows new customers only
	DiscountType             types.String                          `tfsdk:"discount_type"`                // The promotion discount type
	PercentageDiscount       types.Float64                         `tfsdk:"percentage_discount"`          // The promotion discount, percentage
	UnlimitedUses            types.Bool                            `tfsdk:"unlimited_uses"`               // Whether to allow unlimited uses
	UsesAllowed              types.Int32                           `tfsdk:"uses_allowed"`                 // The number of uses allowed by the promotion
	NeverAllowZero           types.Bool                            `tfsdk:"never_allow_zero"`             // Never allow the value of checkout to be zero
	FixedPromotionCode       types.String                          `tfsdk:"fixed_promotion_code"`         // The fixed value for all the promotion codes
	PromotionCodePrefix      types.String                          `tfsdk:"promotion_code_prefix"`        // The prefix for all the codes
	TermDependencyType       types.String                          `tfsdk:"term_dependency_type"`         // The type of dependency to terms
	ApplyToAllBillingPeriods types.Bool                            `tfsdk:"apply_to_all_billing_periods"` // Whether to apply the promotion discount to all billing periods ("TRUE")or the first billing period only ("FALSE")
	CanBeAppliedOnRenewal    types.Bool                            `tfsdk:"can_be_applied_on_renewal"`    // Whether the promotion can be applied on renewal
	BillingPeriodLimit       types.Int32                           `tfsdk:"billing_period_limit"`         // Promotion discount applies to number of billing periods
	FixedDiscountList        []PromotionFixedDiscountResourceModel `tfsdk:"fixed_discount_list"`
	CreateDate               types.Int64                           `tfsdk:"create_date"` // The creation date
	UpdateDate               types.Int64                           `tfsdk:"update_date"` // The update date
}

type PromotionFixedDiscountResourceModel struct {
	FixedDiscountId types.String  `tfsdk:"fixed_discount_id"` // The fixed discount ID
	Currency        types.String  `tfsdk:"currency"`          // The currency of the fixed discount
	Amount          types.String  `tfsdk:"amount"`            // The fixed discount amount
	AmountValue     types.Float64 `tfsdk:"amount_value"`      // The fixed discount amount value
}

func (*PromotionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Promotion represents a special discount. Users can use a promotion code associated with a promotion to get a discount." +
			"For more details, see https://docs.piano.io/promotions/",
		Attributes: map[string]schema.Attribute{
			// always required
			"aid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The application ID",
			},
			// required in request
			"promotion_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The promotion ID",
			},
			// always required
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The promotion name",
			},
			// filled with empty value in create response
			"discount_type": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					// Required on update
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The promotion discount type",
				Validators:          []validator.String{stringvalidator.OneOf("fixed", "percentage")},
			},
			// filled with empty value in create response
			"term_dependency_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The type of dependency to terms.
When the value is "all", the promotion can be applied to app terms.
When the value is "include", the promotion can be applied to those specific terms.
When the value is "unlocked", the promotion allows customers to access special terms that they could not have accessed without the code`,
				Validators: []validator.String{stringvalidator.OneOf("all", "include", "unlocked")},
			},
			// filled with empty value in create response
			"billing_period_limit": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Promotion discount applies to number of billing periods",
			},
			// filled with empty value in create response
			"can_be_applied_on_renewal": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Whether the promotion can be applied on renewal",
			},
			// filled with empty value in create response
			"apply_to_all_billing_periods": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Whether to apply the promotion discount to all billing periods (\"TRUE\")or the first billing period only (\"FALSE\")",
			},
			// filled with empty value in create response
			"never_allow_zero": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Never allow the value of checkout to be zero",
			},
			// filled with empty value in create response
			"fixed_discount_list": schema.ListNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
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
			// filled with empty value in create response
			"new_customers_only": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Whether the promotion allows new customers only",
			},
			// filled with empty value in create response
			"percentage_discount": schema.Float64Attribute{
				// Required when "discount_type" is "percentage"
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The promotion discount, percentage",
			},
			// filled with empty value in create response
			"start_date": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The start date.",
			},
			// filled with empty value in create response
			"end_date": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The end date",
			},
			// nullable in response
			"promotion_code_prefix": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The prefix for all the codes",
			},
			// nullable in response
			"uses_allowed": schema.Int32Attribute{
				Optional: true,
				// updated to null when unlimited_uses = true
				MarkdownDescription: "The number of uses allowed by the promotion. If this value is null, it indicates unlimited uses allowed.",
			},
			// nullable in response
			"fixed_promotion_code": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The fixed value for all the promotion codes",
			},
			// computed
			"create_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The creation date",
			},
			// computed
			"update_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The update date",
			},
			// computed: this value determines the nullability of `use_allowed` field
			"unlimited_uses": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether to allow unlimited uses",
			},
		},
	}
}
func PromotionFixedDiscountResourceModelFrom(data piano_publisher.PromotionFixedDiscount) PromotionFixedDiscountResourceModel {
	ret := PromotionFixedDiscountResourceModel{}
	ret.AmountValue = types.Float64Value(data.AmountValue)
	ret.Amount = types.StringValue(data.Amount)
	ret.Currency = types.StringValue(data.Currency)
	ret.FixedDiscountId = types.StringValue(data.FixedDiscountId)
	return ret
}
func (r *PromotionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PromotionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
	state.UsesAllowed = types.Int32PointerValue(data.UsesAllowed)
	if state.PromotionCodePrefix.IsNull() && data.PromotionCodePrefix != nil && *data.PromotionCodePrefix == "" {
		state.PromotionCodePrefix = types.StringNull()
	} else {
		state.PromotionCodePrefix = types.StringPointerValue(data.PromotionCodePrefix)
	}
	state.PromotionId = types.StringValue(data.PromotionId)
	state.PercentageDiscount = types.Float64Value(data.PercentageDiscount)
	state.NewCustomersOnly = types.BoolValue(data.NewCustomersOnly)
	fixedDiscountListElements := []PromotionFixedDiscountResourceModel{}
	for _, element := range data.FixedDiscountList {
		fixedDiscountListElements = append(fixedDiscountListElements, PromotionFixedDiscountResourceModelFrom(element))
	}
	state.FixedDiscountList = fixedDiscountListElements
	state.EndDate = types.Int64Value(int64(data.EndDate))
	state.NeverAllowZero = types.BoolValue(data.NeverAllowZero)
	state.ApplyToAllBillingPeriods = types.BoolValue(data.ApplyToAllBillingPeriods)
	state.CanBeAppliedOnRenewal = types.BoolValue(data.CanBeAppliedOnRenewal)
	state.BillingPeriodLimit = types.Int32Value(data.BillingPeriodLimit)
	if state.FixedPromotionCode.IsNull() && data.FixedPromotionCode != nil && *data.FixedPromotionCode == "" {
		state.FixedPromotionCode = types.StringNull()
	} else {
		state.FixedPromotionCode = types.StringPointerValue(data.FixedPromotionCode)
	}
	state.Aid = types.StringValue(data.Aid)
	state.TermDependencyType = types.StringValue(string(data.TermDependencyType))
	state.StartDate = types.Int64Value(int64(data.StartDate))
	state.Name = types.StringValue(data.Name)
	state.DiscountType = types.StringValue(string(data.DiscountType))
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *PromotionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state PromotionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	request := piano_publisher.PostPublisherPromotionCreateFormdataRequestBody{
		Aid:                   state.Aid.ValueString(),
		Name:                  state.Name.ValueString(),
		BillingPeriodLimit:    state.BillingPeriodLimit.ValueInt32Pointer(),
		CanBeAppliedOnRenewal: state.ApplyToAllBillingPeriods.ValueBoolPointer(),
		DiscountType:          (*piano_publisher.PostPublisherPromotionCreateRequestDiscountType)(state.DiscountType.ValueStringPointer()),
		NeverAllowZero:        state.NeverAllowZero.ValueBoolPointer(),
		NewCustomersOnly:      *state.NewCustomersOnly.ValueBoolPointer(),
		PromotionCodePrefix:   state.PromotionCodePrefix.ValueStringPointer(),
		TermDependencyType:    (*piano_publisher.PostPublisherPromotionCreateRequestTermDependencyType)(state.TermDependencyType.ValueStringPointer()),
		UsesAllowed:           state.UsesAllowed.ValueInt32Pointer(),
		FixedPromotionCode:    state.FixedPromotionCode.ValueStringPointer(),
	}
	if state.UsesAllowed.IsNull() {
		t := true
		request.UnlimitedUses = &t
	}
	if state.StartDate.ValueInt64Pointer() != nil {
		date := int(state.StartDate.ValueInt64())
		request.StartDate = &date
	}
	if state.EndDate.ValueInt64Pointer() != nil {
		date := int(state.EndDate.ValueInt64())
		request.EndDate = &date
	}
	if state.PercentageDiscount.ValueFloat64Pointer() != nil {
		discount := float32(state.PercentageDiscount.ValueFloat64())
		request.PercentageDiscount = &discount
	}
	response, err := r.client.PostPublisherPromotionCreateWithFormdataBody(ctx, request)
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
	state.UsesAllowed = types.Int32PointerValue(data.UsesAllowed)
	state.PromotionCodePrefix = types.StringPointerValue(data.PromotionCodePrefix)
	state.PromotionId = types.StringValue(data.PromotionId)
	state.UnlimitedUses = types.BoolValue(data.UnlimitedUses)
	state.PercentageDiscount = types.Float64Value(data.PercentageDiscount)
	state.NewCustomersOnly = types.BoolValue(data.NewCustomersOnly)
	fixedDiscountListElements := []PromotionFixedDiscountResourceModel{}
	for _, element := range data.FixedDiscountList {
		fixedDiscountListElements = append(fixedDiscountListElements, PromotionFixedDiscountResourceModelFrom(element))
	}
	state.FixedDiscountList = fixedDiscountListElements
	state.EndDate = types.Int64Value(int64(data.EndDate))
	state.NeverAllowZero = types.BoolValue(data.NeverAllowZero)
	state.ApplyToAllBillingPeriods = types.BoolValue(data.ApplyToAllBillingPeriods)

	state.CanBeAppliedOnRenewal = types.BoolValue(data.CanBeAppliedOnRenewal)
	state.BillingPeriodLimit = types.Int32Value(data.BillingPeriodLimit)
	if state.FixedPromotionCode.IsNull() && data.FixedPromotionCode != nil && *data.FixedPromotionCode == "" {
		state.FixedPromotionCode = types.StringNull()
	} else {
		state.FixedPromotionCode = types.StringPointerValue(data.FixedPromotionCode)
	}
	state.Aid = types.StringValue(data.Aid)
	state.TermDependencyType = types.StringValue(string(data.TermDependencyType))
	state.StartDate = types.Int64Value(int64(data.StartDate))
	state.Name = types.StringValue(data.Name)
	state.DiscountType = types.StringValue(string(data.DiscountType))
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *PromotionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state PromotionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, fmt.Sprintf("DEBUG!!! %#v", state))
	request := piano_publisher.PostPublisherPromotionUpdateFormdataRequestBody{
		Aid:                      state.Aid.ValueString(),
		PromotionId:              state.PromotionId.ValueString(),
		Name:                     state.Name.ValueString(),
		ApplyToAllBillingPeriods: state.ApplyToAllBillingPeriods.ValueBoolPointer(),
		CanBeAppliedOnRenewal:    state.CanBeAppliedOnRenewal.ValueBoolPointer(),
		NeverAllowZero:           state.NeverAllowZero.ValueBoolPointer(),
		UsesAllowed:              state.UsesAllowed.ValueInt32Pointer(),
		BillingPeriodLimit:       state.BillingPeriodLimit.ValueInt32Pointer(),
		DiscountType:             piano_publisher.PostPublisherPromotionUpdateRequestDiscountType(state.DiscountType.ValueString()),
		TermDependencyType:       (*piano_publisher.PostPublisherPromotionUpdateRequestTermDependencyType)(state.TermDependencyType.ValueStringPointer()),
		FixedPromotionCode:       state.FixedPromotionCode.ValueStringPointer(),
		NewCustomersOnly:         state.NewCustomersOnly.ValueBoolPointer(),
		PromotionCodePrefix:      state.PromotionCodePrefix.ValueStringPointer(),
	}
	if state.UsesAllowed.IsNull() {
		t := true
		request.UnlimitedUses = &t
	}
	if state.StartDate.ValueInt64Pointer() != nil {
		date := int(state.StartDate.ValueInt64())
		request.StartDate = &date
	}
	if state.EndDate.ValueInt64Pointer() != nil {
		date := int(state.EndDate.ValueInt64())
		request.EndDate = &date
	}
	if state.PercentageDiscount.ValueFloat64Pointer() != nil {
		discount := float32(state.PercentageDiscount.ValueFloat64())
		request.PercentageDiscount = &discount
	}
	response, err := r.client.PostPublisherPromotionUpdateWithFormdataBody(ctx, request)
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
	state.UsesAllowed = types.Int32PointerValue(data.UsesAllowed)
	if state.PromotionCodePrefix.IsNull() && data.PromotionCodePrefix != nil && *data.PromotionCodePrefix == "" {
		state.PromotionCodePrefix = types.StringNull()
	} else {
		state.PromotionCodePrefix = types.StringPointerValue(data.PromotionCodePrefix)
	}
	state.PromotionId = types.StringValue(data.PromotionId)
	state.UnlimitedUses = types.BoolValue(data.UnlimitedUses)
	state.PercentageDiscount = types.Float64Value(data.PercentageDiscount)
	state.NewCustomersOnly = types.BoolValue(data.NewCustomersOnly)
	fixedDiscountListElements := []PromotionFixedDiscountResourceModel{}
	for _, element := range data.FixedDiscountList {
		fixedDiscountListElements = append(fixedDiscountListElements, PromotionFixedDiscountResourceModelFrom(element))
	}
	state.FixedDiscountList = fixedDiscountListElements
	state.EndDate = types.Int64Value(int64(data.EndDate))
	state.NeverAllowZero = types.BoolValue(data.NeverAllowZero)
	state.ApplyToAllBillingPeriods = types.BoolValue(data.ApplyToAllBillingPeriods)
	state.CanBeAppliedOnRenewal = types.BoolValue(data.CanBeAppliedOnRenewal)
	state.BillingPeriodLimit = types.Int32Value(data.BillingPeriodLimit)
	state.FixedPromotionCode = types.StringPointerValue(data.FixedPromotionCode)
	state.Aid = types.StringValue(data.Aid)
	state.TermDependencyType = types.StringValue(string(data.TermDependencyType))
	state.StartDate = types.Int64Value(int64(data.StartDate))
	state.Name = types.StringValue(data.Name)
	state.DiscountType = types.StringValue(string(data.DiscountType))
	state.CreateDate = types.Int64Value(int64(data.CreateDate))
	state.UpdateDate = types.Int64Value(int64(data.UpdateDate))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *PromotionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PromotionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, fmt.Sprintf("deleting promotion %s:%s in $%s", state.Name.ValueString(), state.PromotionId.ValueString(), state.Aid.ValueString()))
	response, err := r.client.PostPublisherPromotionDeleteWithFormdataBody(ctx, piano_publisher.PostPublisherPromotionDeleteFormdataRequestBody{
		Aid:         state.Aid.ValueString(),
		PromotionId: state.PromotionId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete licensee, got error: %s", err))
		return
	}
	_, err = syntax.AnyResponseFrom(response, &resp.Diagnostics)
	// TODO: handle 3009 -- Can not delete promotion with claimed codes
	if err != nil {
		return
	}
}
func (*PromotionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	promotionId, err := PromotionIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource resource id", fmt.Sprintf("Unable to parse promotion id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), promotionId.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("promotion_id"), promotionId.PromotionId)...)
}

// PromotionId represents a piano.io promotion resource identifier in "{aid}/{promotion_id}" format.
type PromotionId struct {
	Aid         string
	PromotionId string
}

func PromotionIdFromString(input string) (*PromotionId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, errors.New("resource resource id must be in {aid}/{rid} format")
	}
	data := PromotionId{
		Aid:         parts[0],
		PromotionId: parts[1],
	}
	return &data, nil
}
