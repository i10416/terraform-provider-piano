// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.ListTypable = ExternalAPIFieldResourceModelList{}
var _ basetypes.ListValuable = ExternalAPIFieldResourceModelListValue{}

type ExternalAPIFieldResourceModelList struct {
	basetypes.ListType
}

func (l ExternalAPIFieldResourceModelList) String() string {
	return fmt.Sprintf("ExternalAPIFieldResourceModelList")
}
func (l ExternalAPIFieldResourceModelList) Equal(a attr.Type) bool {
	r, ok := a.(ExternalAPIFieldResourceModelList)
	if !ok {
		return false
	}
	return l.ListType.Equal(r.ListType)
}

func (l ExternalAPIFieldResourceModelList) ValueFromList(ctx context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	if in.IsNull() {
		return ExternalAPIFieldResourceModelListValueNull(), nil
	}
	if in.IsUnknown() {
		return ExternalAPIFieldResourceModelListValueUnknown(), nil
	}
	listValue, diags := basetypes.NewListValue(ExternalAPIFieldAttrType(), in.Elements())
	if diags.HasError() {
		return ExternalAPIFieldResourceModelListValueUnknown(), diags
	}
	return ExternalAPIFieldResourceModelListValue{ListValue: listValue}, nil
}

func (l ExternalAPIFieldResourceModelList) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := l.ListType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	listValue, ok := attrValue.(basetypes.ListValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}
	listValuable, diags := l.ValueFromList(ctx, listValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting ListValue to ListValuable: %v", diags)
	}
	return listValuable, nil
}
func (l ExternalAPIFieldResourceModelList) ValueType(context.Context) attr.Value {
	return ExternalAPIFieldResourceModelListValue{}
}

type ExternalAPIFieldResourceModelListValue struct {
	basetypes.ListValue
}

func ExternalAPIFieldAttrType() attr.Type {
	return basetypes.ObjectType{
		AttrTypes: map[string]attr.Type{
			"mandatory":     types.BoolType,
			"description":   types.StringType,
			"hidden":        types.BoolType,
			"order":         types.Int32Type,
			"default_value": types.StringType,
			"field_title":   types.StringType,
			"type":          types.StringType,
			"field_name":    types.StringType,
			"editable":      types.StringType,
		},
	}
}

func ExternalAPIFieldResourceModelListValueUnknown() ExternalAPIFieldResourceModelListValue {
	return ExternalAPIFieldResourceModelListValue{ListValue: types.ListUnknown(ExternalAPIFieldAttrType())}
}
func ExternalAPIFieldResourceModelListValueNull() ExternalAPIFieldResourceModelListValue {
	return ExternalAPIFieldResourceModelListValue{ListValue: basetypes.NewListNull(
		ExternalAPIFieldAttrType(),
	)}
}
func (v ExternalAPIFieldResourceModelListValue) Equal(o attr.Value) bool {
	other, ok := o.(ExternalAPIFieldResourceModelListValue)
	if !ok {
		return false
	}
	return v.ListValue.Equal(other.ListValue)
}

func (v ExternalAPIFieldResourceModelListValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tv := v.ListValue
	if tv.ElementType(ctx) == nil {
		tv = ExternalAPIFieldResourceModelListValueNull().ListValue
	}
	return tv.ToTerraformValue(ctx)
}

func (v ExternalAPIFieldResourceModelListValue) Type(ctx context.Context) attr.Type {
	return ExternalAPIFieldResourceModelList{
		ListType: basetypes.ListType{
			ElemType: ExternalAPIFieldAttrType(),
		},
	}
}

func (v ExternalAPIFieldResourceModelListValue) Value(ctx context.Context) ([]attr.Value, diag.Diagnostics) {
	return v.ListValue.Elements(), nil
}
