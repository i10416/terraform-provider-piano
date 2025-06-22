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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &ResourceResource{}
	_ resource.ResourceWithImportState = &ResourceResource{}
)

func NewResourceResource() resource.Resource {
	return &ResourceResource{}
}

// ResourceResource defines the resource implementation.
type ResourceResource struct {
	client *piano_publisher.Client
}

// ResourceResourceModel describes the resource model.
type ResourceResourceModel struct {
	Rid            types.String `tfsdk:"rid"`              // The resource ID
	Aid            types.String `tfsdk:"aid"`              // The application ID
	Deleted        types.Bool   `tfsdk:"deleted"`          // Whether the object is deleted
	Disabled       types.Bool   `tfsdk:"disabled"`         // Whether the object is disabled
	CreateDate     types.Int64  `tfsdk:"create_date"`      // The creation date
	UpdateDate     types.Int64  `tfsdk:"update_date"`      // The update date
	PublishDate    types.Int64  `tfsdk:"publish_date"`     // The publish date
	Name           types.String `tfsdk:"name"`             // The name
	Description    types.String `tfsdk:"description"`      // The resource description
	ImageUrl       types.String `tfsdk:"image_url"`        // The URL of the resource image
	Type           types.String `tfsdk:"type"`             // The type of the resource (0: Standard, 4: Bundle)
	BundleType     types.String `tfsdk:"bundle_type"`      // The resource bundle type
	PurchaseUrl    types.String `tfsdk:"purchase_url"`     // The URL of the purchase page
	ResourceUrl    types.String `tfsdk:"resource_url"`     // The URL of the resource
	ExternalId     types.String `tfsdk:"external_id"`      // The external ID; defined by the client
	IsFbiaResource types.Bool   `tfsdk:"is_fbia_resource"` // Enable the resource for Facebook Subscriptions in Instant Articles
}

func (r *ResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *ResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource resource. Resources are fundamental concept used to control access to " +
			"content you’re gating (e.g. an article, a movie, a blog post, a pdf, access to a forum, access to premium site content, etc.) in piano.io.",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				MarkdownDescription: "The application ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name",
				Required:            true,
			},
			"rid": schema.StringAttribute{
				MarkdownDescription: "The resource ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The resource description",
				Optional:            true,
			},
			"deleted": schema.BoolAttribute{
				MarkdownDescription: "Whether the object is deleted",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the object is disabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"create_date": schema.Int64Attribute{
				MarkdownDescription: "The creation date timestamp",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"update_date": schema.Int64Attribute{
				MarkdownDescription: "The update date timestamp",
				Computed:            true,
			},
			"publish_date": schema.Int64Attribute{
				MarkdownDescription: "The publish date timestamp",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"image_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the resource image",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource (0: Standard, 4: Bundle)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bundle_type": schema.StringAttribute{
				MarkdownDescription: "The resource bundle type",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"purchase_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the purchase page",
				Optional:            true,
			},
			"resource_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the resource",
				Optional:            true,
			},
			"external_id": schema.StringAttribute{
				MarkdownDescription: "The external ID; defined by the client",
				Optional:            true,
			},
			"is_fbia_resource": schema.BoolAttribute{
				MarkdownDescription: "Enable the resource for Facebook Subscriptions in Instant Articles",
				Optional:            true,
			},
		},
	}
}

func (r *ResourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*PianoProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *piano_publisher.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = &client.publisherClient
}

func (r *ResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state ResourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("creating resource %s in %s", state.Name.ValueString(), state.Aid.ValueString()))

	response, err := r.client.PostPublisherResourceCreateWithFormdataBody(ctx, piano_publisher.PostPublisherResourceCreateFormdataRequestBody{
		Aid:         state.Aid.ValueString(),
		Name:        state.Name.ValueString(),
		Description: state.Description.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.ResourceResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}
	// Computed, ReadOnly
	state.Rid = types.StringValue(result.Resource.Rid)
	state.CreateDate = types.Int64Value(int64(result.Resource.CreateDate))
	state.UpdateDate = types.Int64Value(int64(result.Resource.UpdateDate))
	state.PublishDate = types.Int64Value(int64(result.Resource.PublishDate))
	state.Deleted = types.BoolValue(result.Resource.Deleted)
	state.Type = types.StringValue(string(result.Resource.Type))
	state.BundleType = types.StringPointerValue((*string)(result.Resource.BundleType))
	// Updatable
	state.Name = types.StringValue(result.Resource.Name)
	if state.Description.IsNull() && result.Resource.Description != nil && *result.Resource.Description == "" {
		result.Resource.Description = nil
	}
	state.Description = types.StringPointerValue(result.Resource.Description)
	state.ExternalId = types.StringPointerValue(result.Resource.ExternalId)
	state.ImageUrl = types.StringPointerValue(result.Resource.ImageUrl)
	state.ResourceUrl = types.StringPointerValue(result.Resource.ResourceUrl)
	state.IsFbiaResource = types.BoolValue(result.Resource.IsFbiaResource)
	state.Disabled = types.BoolValue(result.Resource.Disabled)
	// Not-Updatable
	state.PurchaseUrl = types.StringPointerValue(result.Resource.PurchaseUrl)

	tflog.Info(ctx, fmt.Sprintf("complete creating resource %s(id: %s)", state.Name, state.Rid))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.GetPublisherResourceGet(ctx, &piano_publisher.GetPublisherResourceGetParams{
		Aid: state.Aid.ValueString(),
		Rid: state.Rid.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch resource, got error: %s", err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.ResourceResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	// Computed, ReadOnly
	state.CreateDate = types.Int64Value(int64(result.Resource.CreateDate))
	state.UpdateDate = types.Int64Value(int64(result.Resource.UpdateDate))
	state.PublishDate = types.Int64Value(int64(result.Resource.PublishDate))
	state.Deleted = types.BoolValue(result.Resource.Deleted)
	state.Type = types.StringValue(string(result.Resource.Type))
	state.BundleType = types.StringPointerValue((*string)(result.Resource.BundleType))
	// Updatable
	state.Name = types.StringValue(result.Resource.Name)
	if state.Description.IsNull() && result.Resource.Description != nil && *result.Resource.Description == "" {
		result.Resource.Description = nil
	}
	state.Description = types.StringPointerValue(result.Resource.Description)
	state.ExternalId = types.StringPointerValue(result.Resource.ExternalId)
	state.ImageUrl = types.StringPointerValue(result.Resource.ImageUrl)
	state.ResourceUrl = types.StringPointerValue(result.Resource.ResourceUrl)
	state.IsFbiaResource = types.BoolValue(result.Resource.IsFbiaResource)
	state.Disabled = types.BoolValue(result.Resource.Disabled)
	// Not-Updatable
	state.PurchaseUrl = types.StringPointerValue(result.Resource.PurchaseUrl)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// NOTE: This state contains only updated values at first
	var state ResourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("updating resource %s(id:%s) in %s", state.Name.ValueString(), state.Rid.ValueString(), state.Aid.ValueString()))
	request := piano_publisher.PostPublisherResourceUpdateFormdataRequestBody{
		Aid:            state.Aid.ValueString(),
		Rid:            state.Rid.ValueString(),
		Name:           state.Name.ValueStringPointer(),
		Description:    state.Description.ValueStringPointer(),
		Disabled:       state.Disabled.ValueBoolPointer(),
		ExternalId:     state.ExternalId.ValueStringPointer(),
		ImageUrl:       state.ImageUrl.ValueStringPointer(),
		IsFbiaResource: state.IsFbiaResource.ValueBoolPointer(),
		ResourceUrl:    state.ResourceUrl.ValueStringPointer(),
	}

	response, err := r.client.PostPublisherResourceUpdateWithFormdataBody(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update resource, got error: %s", err))
		tflog.Error(ctx, fmt.Sprintf("Unable to update resource(%s): %e", state.Rid, err))
		return
	}
	anyResponse, err := syntax.SuccessfulResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.ResourceResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	// Computed, ReadOnly
	state.CreateDate = types.Int64Value(int64(result.Resource.CreateDate))
	state.UpdateDate = types.Int64Value(int64(result.Resource.UpdateDate))
	state.PublishDate = types.Int64Value(int64(result.Resource.PublishDate))
	state.Deleted = types.BoolValue(result.Resource.Deleted)
	state.Type = types.StringValue(string(result.Resource.Type))
	state.BundleType = types.StringPointerValue((*string)(result.Resource.BundleType))
	// Updatable
	state.Name = types.StringValue(result.Resource.Name)
	if state.Description.IsNull() && result.Resource.Description != nil && *result.Resource.Description == "" {
		result.Resource.Description = nil
	}
	state.Description = types.StringPointerValue(result.Resource.Description)
	state.ExternalId = types.StringPointerValue(result.Resource.ExternalId)
	state.ImageUrl = types.StringPointerValue(result.Resource.ImageUrl)
	state.ResourceUrl = types.StringPointerValue(result.Resource.ResourceUrl)
	state.IsFbiaResource = types.BoolValue(result.Resource.IsFbiaResource)
	state.Disabled = types.BoolValue(result.Resource.Disabled)
	// Not-Updatable
	state.PurchaseUrl = types.StringPointerValue(result.Resource.PurchaseUrl)

	tflog.Info(ctx, fmt.Sprintf("complete updating resource %s(id: %s)", state.Name, state.Rid))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("deleting Resource %s:%s in $%s", state.Name.ValueString(), state.Rid.ValueString(), state.Aid.ValueString()))
	response, err := r.client.PostPublisherResourceDeleteWithFormdataBody(ctx, piano_publisher.PostPublisherResourceDeleteFormdataRequestBody{
		Aid: state.Aid.ValueString(),
		Rid: state.Rid.ValueString(),
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

func (r *ResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceId, err := ResourceResourceIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource resource id", fmt.Sprintf("Unable to parse resource resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), resourceId.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rid"), resourceId.ResourceId)...)
}

func ResourceManagerUidsStringFromModels(models []ManagerResourceModel) string {
	managerUids := []string{}
	for _, m := range models {
		managerUids = append(managerUids, m.UID.ValueString())
	}
	managerUidsAsString := strings.Join(managerUids, ",")
	return managerUidsAsString
}

// ResourceResourceId represents a piano.io Resource resource identifier in "{aid}/{rid}" format.
type ResourceResourceId struct {
	Aid        string
	ResourceId string
}

func ResourceResourceIdFromString(input string) (*ResourceResourceId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, errors.New("resource resource id must be in {aid}/{rid} format")
	}
	data := ResourceResourceId{
		Aid:        parts[0],
		ResourceId: parts[1],
	}
	return &data, nil
}
