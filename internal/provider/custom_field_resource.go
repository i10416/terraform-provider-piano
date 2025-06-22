// // Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// // SPDX-License-Identifier: MPL-2.0
package provider

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"strings"
// 	"terraform-provider-piano/internal/piano_id"
// 	"terraform-provider-piano/internal/piano_publisher"
// 	"terraform-provider-piano/internal/syntax"

// 	"github.com/hashicorp/terraform-plugin-framework/path"
// 	"github.com/hashicorp/terraform-plugin-framework/resource"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
// 	"github.com/hashicorp/terraform-plugin-framework/types"
// 	"github.com/hashicorp/terraform-plugin-log/tflog"
// )

// var (
// 	_ resource.Resource                = &CustomFieldResource{}
// 	_ resource.ResourceWithImportState = &CustomFieldResource{}
// )

// type CustomFieldResource struct {
// 	client *piano_id.Client
// }

// func NewCustomFieldResource() resource.Resource {
// 	return &CustomFieldResource{}
// }
// func (r *CustomFieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
// 	if req.ProviderData == nil {
// 		return
// 	}
// 	client, ok := req.ProviderData.(PianoProviderData)

// 	if !ok {
// 		resp.Diagnostics.AddError(
// 			"Unexpected Resource Configure Type",
// 			fmt.Sprintf("Expected PianoProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
// 		)

// 		return
// 	}

// 	r.client = &client.idClient
// }

// func (r *CustomFieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
// 	resp.TypeName = req.ProviderTypeName + "_offer"
// }

// type CustomFieldResourceModel struct {
// 	Name          types.String `tfsdk:"name"`     // The offer name
// 	Aid           types.String `tfsdk:"aid"`      // The application ID
// 	CustomFieldId types.String `tfsdk:"offer_id"` // The offer ID
// }

// func (*CustomFieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
// 	resp.Schema = schema.Schema{
// 		Attributes: map[string]schema.Attribute{
// 			"aid": schema.StringAttribute{
// 				Required:            true,
// 				MarkdownDescription: "The application ID",
// 			},
// 			"offer_id": schema.StringAttribute{
// 				Computed: true,
// 				PlanModifiers: []planmodifier.String{
// 					stringplanmodifier.UseStateForUnknown(),
// 				},
// 				MarkdownDescription: "The offer ID",
// 			},
// 			"name": schema.StringAttribute{
// 				Required:            true,
// 				MarkdownDescription: "The offer name",
// 			},
// 		},
// 	}
// }

// func (r *CustomFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
// 	var state CustomFieldResourceModel
// 	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// 	response, err := r.client.GetPublisherCustomFieldGet(ctx, &piano_publisher.GetPublisherCustomFieldGetParams{
// 		Aid:           state.Aid.ValueString(),
// 		CustomFieldId: state.CustomFieldId.ValueString(),
// 	})
// 	if err != nil {
// 		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch promotion, got error: %s", err))
// 		return
// 	}
// 	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
// 	if err != nil {
// 		return
// 	}

// 	result := piano_publisher.CustomFieldModelResult{}
// 	err = json.Unmarshal(anyResponse.Raw, &result)
// 	if err != nil {
// 		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
// 		return
// 	}

// 	data := result.CustomField
// 	state.CustomFieldId = types.StringValue(data.CustomFieldId)
// 	state.Aid = types.StringValue(data.Aid)

// 	state.Name = types.StringValue(data.Name)
// 	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
// }
// func (r *CustomFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
// 	var state CustomFieldResourceModel
// 	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// 	response, err := r.client.PostPublisherCustomFieldCreateWithFormdataBody(ctx, piano_publisher.PostPublisherCustomFieldCreateFormdataRequestBody{
// 		Aid:  state.Aid.ValueString(),
// 		Name: state.Name.ValueString(),
// 	})
// 	if err != nil {
// 		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create offer, got error: %s", err))
// 		return
// 	}
// 	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
// 	if err != nil {
// 		return
// 	}

// 	result := piano_publisher.CustomFieldModelResult{}
// 	err = json.Unmarshal(anyResponse.Raw, &result)
// 	if err != nil {
// 		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %e", err))
// 		return
// 	}

// 	data := result.CustomField

// 	state.CustomFieldId = types.StringValue(data.CustomFieldId)
// 	state.Aid = types.StringValue(data.Aid)

// 	state.Name = types.StringValue(data.Name)
// 	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
// }
// func (r *CustomFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
// 	var state CustomFieldResourceModel
// 	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// 	response, err := r.client.PostPublisherCustomFieldUpdateWithFormdataBody(ctx, piano_publisher.PostPublisherCustomFieldUpdateFormdataRequestBody{
// 		Aid:  state.Aid.ValueString(),
// 		Name: state.Name.ValueString(),
// 	})
// 	if err != nil {
// 		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update offer, got error: %s", err))
// 		return
// 	}
// 	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
// 	if err != nil {
// 		return
// 	}

// 	result := piano_publisher.CustomFieldModelResult{}
// 	err = json.Unmarshal(anyResponse.Raw, &result)
// 	if err != nil {
// 		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
// 		return
// 	}

// 	data := result.CustomField
// 	state.CustomFieldId = types.StringValue(data.CustomFieldId)
// 	state.Aid = types.StringValue(data.Aid)
// 	state.Name = types.StringValue(data.Name)
// 	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
// }
// func (r *CustomFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
// 	var state CustomFieldResourceModel
// 	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// 	tflog.Info(ctx, fmt.Sprintf("deleting promotion %s:%s in $%s", state.Name.ValueString(), state.CustomFieldId.ValueString(), state.Aid.ValueString()))
// 	response, err := r.client.PostPublisherCustomFieldDeleteWithFormdataBody(ctx, piano_publisher.PostPublisherCustomFieldDeleteFormdataRequestBody{
// 		Aid:           state.Aid.ValueString(),
// 		CustomFieldId: state.CustomFieldId.ValueString(),
// 	})
// 	if err != nil {
// 		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete licensee, got error: %s", err))
// 		return
// 	}
// 	_, err = syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
// 	if err != nil {
// 		return
// 	}
// }
// func (r *CustomFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	id, err := CustomFieldIdFromString(req.ID)
// 	if err != nil {
// 		resp.Diagnostics.AddError("Invalid offer resource id", fmt.Sprintf("Unable to parse offer resource id, got error: %s", err))
// 		return
// 	}
// 	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), id.Aid)...)
// 	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("offer_id"), id.CustomFieldId)...)
// }

// // CustomFieldId represents a piano.io promotion resource identifier in "{aid}/{offer_id}" format.
// type CustomFieldResourceId struct {
// 	Aid           string
// 	CustomFieldId string
// }

// func CustomFieldIdFromString(input string) (*CustomFieldResourceId, error) {
// 	parts := strings.Split(input, "/")
// 	if len(parts) != 2 {
// 		return nil, errors.New("offer resource id must be in {aid}/{offer_id} format")
// 	}
// 	data := CustomFieldResourceId{
// 		Aid:           parts[0],
// 		CustomFieldId: parts[1],
// 	}
// 	return &data, nil
// }
