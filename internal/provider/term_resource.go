// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ScheduleResourceModel struct {
	Aid        types.String `tfsdk:"aid"`         // The application ID
	CreateDate types.Int64  `tfsdk:"create_date"` // The creation date
	Deleted    types.Bool   `tfsdk:"deleted"`     // Whether the object is deleted
	Name       types.String `tfsdk:"name"`        // The schedule name
	ScheduleId types.String `tfsdk:"schedule_id"` // The schedule ID
	UpdateDate types.Int64  `tfsdk:"update_date"` // The update date
}

type VoucheringPolicyResourceModel struct {
	VoucheringPolicyBillingPlan            types.String `tfsdk:"vouchering_policy_billing_plan"`             // The billing plan of the vouchering policy
	VoucheringPolicyBillingPlanDescription types.String `tfsdk:"vouchering_policy_billing_plan_description"` // The description of the vouchering policy billing plan
	VoucheringPolicyId                     types.String `tfsdk:"vouchering_policy_id"`                       // The vouchering policy ID
	VoucheringPolicyRedemptionUrl          types.String `tfsdk:"vouchering_policy_redemption_url"`           // The vouchering policy redemption URL
}
type LightOfferResourceModel struct {
	Name    types.String `tfsdk:"name"`     // The offer name
	OfferId types.String `tfsdk:"offer_id"` // The offer ID
}

// TermResourceId represents a piano.io contract resource identifier in "{aid}/{contract_id}" format.
type TermResourceId struct {
	Aid    string
	TermId string
}

func TermResourceIdFromString(input string) (*TermResourceId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, errors.New("Term resource id must be in {aid}/{term_id} format")
	}
	return &TermResourceId{Aid: parts[0], TermId: parts[1]}, nil
}
