// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-piano/internal/piano_publisher"
	"terraform-provider-piano/internal/syntax"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ResourceDataSource{}
	_ datasource.DataSourceWithConfigure = &ResourceDataSource{}
)

func NewResourceDataSource() datasource.DataSource {
	return &ResourceDataSource{}
}

// ResourceDataSource defines the data source implementation.
type ResourceDataSource struct {
	client *piano_publisher.Client
}

// ResourceDataSourceModel describes the data source data model.
type ResourceDataSourceModel struct {
	Rid             types.String `tfsdk:"rid"`               // The resource ID
	Aid             types.String `tfsdk:"aid"`               // The application ID
	Deleted         types.Bool   `tfsdk:"deleted"`           // Whether the object is deleted
	Disabled        types.Bool   `tfsdk:"disabled"`          // Whether the object is disabled
	CreateDate      types.Int64  `tfsdk:"create_date"`       // The creation date
	UpdateDate      types.Int64  `tfsdk:"update_date"`       // The update date
	PublishDate     types.Int64  `tfsdk:"publish_date"`      // The publish date
	Name            types.String `tfsdk:"name"`              // The name
	Description     types.String `tfsdk:"description"`       // The resource description
	ImageUrl        types.String `tfsdk:"image_url"`         // The URL of the resource image
	Type            types.String `tfsdk:"type"`              // The type of the resource (0: Standard, 4: Bundle)
	TypeLabel       types.String `tfsdk:"type_label"`        // The resource type label ("Standard" or "Bundle")
	BundleType      types.String `tfsdk:"bundle_type"`       // The resource bundle type
	BundleTypeLabel types.String `tfsdk:"bundle_type_label"` // The bundle type label
	PurchaseUrl     types.String `tfsdk:"purchase_url"`      // The URL of the purchase page
	ResourceUrl     types.String `tfsdk:"resource_url"`      // The URL of the resource
	ExternalId      types.String `tfsdk:"external_id"`       // The external ID; defined by the client
	IsFbiaResource  types.Bool   `tfsdk:"is_fbia_resource"`  // Enable the resource for Facebook Subscriptions in Instant Articles
}

func (d *ResourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (d *ResourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource data source. This data source is used to get the resource details.",
		Attributes: map[string]schema.Attribute{
			"rid": schema.StringAttribute{
				MarkdownDescription: "The resource ID",
				Required:            true,
			},
			"aid": schema.StringAttribute{
				MarkdownDescription: "The application ID",
				Required:            true,
			},
			"deleted": schema.BoolAttribute{
				MarkdownDescription: "Whether the object is deleted",
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the object is disabled",
				Computed:            true,
			},
			"create_date": schema.Int64Attribute{
				MarkdownDescription: "The creation date timestamp",
				Computed:            true,
			},
			"update_date": schema.Int64Attribute{
				MarkdownDescription: "The update date timestamp",
				Computed:            true,
			},
			"publish_date": schema.Int64Attribute{
				MarkdownDescription: "The publish date timestamp",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The resource description",
				Computed:            true,
			},
			"image_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the resource image",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the resource (0: Standard, 4: Bundle)",
				Computed:            true,
			},
			"type_label": schema.StringAttribute{
				MarkdownDescription: "The resource type label ('Standard' or 'Bundle')",
				Computed:            true,
			},
			"bundle_type": schema.StringAttribute{
				MarkdownDescription: "The resource bundle type",
				Computed:            true,
			},
			"bundle_type_label": schema.StringAttribute{
				MarkdownDescription: "The bundle type label",
				Computed:            true,
			},
			"purchase_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the purchase page",
				Computed:            true,
			},
			"resource_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the resource",
				Computed:            true,
			},
			"external_id": schema.StringAttribute{
				MarkdownDescription: "The external ID; defined by the client",
				Computed:            true,
			},
			"is_fbia_resource": schema.BoolAttribute{
				MarkdownDescription: "Enable the resource for Facebook Subscriptions in Instant Articles",
				Computed:            true,
			},
		},
	}
}

func (d *ResourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *ResourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ResourceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := d.client.GetPublisherResourceGet(ctx, &piano_publisher.GetPublisherResourceGetParams{
		Aid: data.Aid.ValueString(),
		Rid: data.Rid.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch licensee, got error: %s", err))
		return
	}
	anyResponse, err := syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.ResourceResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	data.Name = types.StringValue(result.Resource.Name)
	data.Type = types.StringValue(string(result.Resource.Type))
	data.TypeLabel = types.StringValue(string(result.Resource.TypeLabel))
	data.BundleType = types.StringPointerValue((*string)(result.Resource.BundleType))
	data.BundleTypeLabel = types.StringPointerValue((*string)(result.Resource.BundleTypeLabel))
	data.Description = types.StringPointerValue(result.Resource.Description)
	data.CreateDate = types.Int64Value(int64(result.Resource.CreateDate))
	data.UpdateDate = types.Int64Value(int64(result.Resource.UpdateDate))
	data.PublishDate = types.Int64Value(int64(result.Resource.PublishDate))
	data.ExternalId = types.StringPointerValue(result.Resource.ExternalId)
	data.Deleted = types.BoolValue(result.Resource.Deleted)
	data.Disabled = types.BoolValue(result.Resource.Disabled)
	data.ImageUrl = types.StringPointerValue(result.Resource.ImageUrl)
	data.PurchaseUrl = types.StringPointerValue(result.Resource.PurchaseUrl)
	data.ResourceUrl = types.StringPointerValue(result.Resource.ResourceUrl)
	data.IsFbiaResource = types.BoolValue(result.Resource.IsFbiaResource)
	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
