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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &LicenseeResource{}
	_ resource.ResourceWithImportState = &LicenseeResource{}
)

func NewLicenseeResource() resource.Resource {
	return &LicenseeResource{}
}

// LicenseeResource defines the resource implementation.
type LicenseeResource struct {
	client *piano_publisher.Client
}

// LicenseeResourceModel describes the resource model.
type LicenseeResourceModel struct {
	// required
	Aid        types.String `tfsdk:"aid"`
	Name       types.String `tfsdk:"name"`
	LicenseeId types.String `tfsdk:"licensee_id"`
	// optional
	Description     types.String                  `tfsdk:"description"`
	LogoUrl         types.String                  `tfsdk:"logo_url"`
	Representatives []RepresentativeResourceModel `tfsdk:"representatives"`
	Managers        []ManagerResourceModel        `tfsdk:"managers"`
}

type RepresentativeResourceModel struct {
	Email types.String `tfsdk:"email"`
}

type ManagerResourceModel struct {
	UID          types.String `tfsdk:"uid"`           // The user's ID
	FirstName    types.String `tfsdk:"first_name"`    // The user's first name
	LastName     types.String `tfsdk:"last_name"`     // The user's last name
	PersonalName types.String `tfsdk:"personal_name"` // The user's personal name. Name and surname ordered as per locale
}

func (r *LicenseeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_licensee"
}

func (r *LicenseeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Licensee resource",
		Attributes: map[string]schema.Attribute{
			"aid": schema.StringAttribute{
				MarkdownDescription: "piano application id",
				Required:            true,
			},
			"licensee_id": schema.StringAttribute{
				MarkdownDescription: "The public ID of the licensee",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the licensee. This will be displayed to end-users in an email and on-site messaging.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the licensee",
				Optional:            true,
			},
			"logo_url": schema.StringAttribute{
				Description:         "An image that can be used to represent the logo for the licensee. This allows specific elements to be branded within creatives.",
				MarkdownDescription: "A relative URL of the licensee's logo",
				Optional:            true,
			},
			"managers": schema.ListNestedAttribute{
				Sensitive:   true,
				Required:    true,
				Description: "This is the person/people from your team who will be responsible for maintaining this licensee and relationship",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uid": schema.StringAttribute{
							Required: true,
						},
						"first_name": schema.StringAttribute{
							Required: true,
						},
						"last_name": schema.StringAttribute{
							Required: true,
						},
						"personal_name": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"representatives": schema.ListNestedAttribute{
				Required:  true,
				Sensitive: true,
				MarkdownDescription: "This is the person from the licensee group " +
					"that is responsible for managing the relationship within their organization. " +
					"This group has elevated permissions to modify licenses within their group.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *LicenseeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*piano_publisher.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *piano_publisher.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *LicenseeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LicenseeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	managerIdsAsString := LicenseeManagerUidsStringFromModels(plan.Managers)

	representativesAsString, err := LicenseeRepresentativesStringFromModels(plan.Representatives)
	if err != nil {
		resp.Diagnostics.AddError("Encoding Error", fmt.Sprintf("Unable to encode representatives as string: %v", plan.Representatives))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("creating licensee %s in %s", plan.Name.ValueString(), plan.Aid.ValueString()))

	response, err := r.client.PostPublisherLicensingLicenseeCreateWithFormdataBody(ctx, piano_publisher.PostPublisherLicensingLicenseeCreateFormdataRequestBody{
		Aid:             plan.Aid.ValueString(),
		Name:            plan.Name.ValueString(),
		ManagerUids:     managerIdsAsString,
		Representatives: representativesAsString,
		Description:     plan.Description.ValueStringPointer(),
		LogoUrl:         plan.LogoUrl.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}
	anyResponse, err := syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.LicenseeResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}
	plan.LicenseeId = types.StringValue(result.Licensee.LicenseeId)

	plan.Name = types.StringValue(result.Licensee.Name)
	plan.Description = types.StringPointerValue(result.Licensee.Description)
	plan.LogoUrl = types.StringPointerValue(result.Licensee.LogoUrl)

	managers := ManagerModelsFromData(result.Licensee.Managers)

	plan.Managers = managers

	representatives := []RepresentativeResourceModel{}
	for _, r := range result.Licensee.Representatives {
		rep := RepresentativeResourceModel{}
		rep.Email = types.StringValue(r.Email)
		representatives = append(representatives, rep)
	}
	plan.Representatives = representatives
	tflog.Info(ctx, fmt.Sprintf("complete creating licensee %s(id: %s)", plan.Name, plan.LicenseeId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LicenseeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LicenseeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.GetPublisherLicensingLicenseeGet(ctx, &piano_publisher.GetPublisherLicensingLicenseeGetParams{
		Aid:        state.Aid.ValueString(),
		LicenseeId: state.LicenseeId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch licensee, got error: %s", err))
		return
	}
	anyResponse, err := syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.LicenseeResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	state.Name = types.StringValue(result.Licensee.Name)
	state.Description = types.StringPointerValue(result.Licensee.Description)
	state.LogoUrl = types.StringPointerValue(result.Licensee.LogoUrl)

	managers := ManagerModelsFromData(result.Licensee.Managers)

	state.Managers = managers

	representatives := []RepresentativeResourceModel{}
	for _, r := range result.Licensee.Representatives {
		rep := RepresentativeResourceModel{}
		rep.Email = types.StringValue(r.Email)
		representatives = append(representatives, rep)
	}
	state.Representatives = representatives

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LicenseeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LicenseeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("updating licensee %s(id:%s) in %s", plan.Name.ValueString(), plan.LicenseeId.ValueString(), plan.Aid.ValueString()))

	representativesAsString, err := LicenseeRepresentativesStringFromModels(plan.Representatives)
	if err != nil {
		resp.Diagnostics.AddError("Encoding Error", fmt.Sprintf("Unable to encode representatives as string: %v", plan.Representatives))
		return
	}

	managerUidsAsString := LicenseeManagerUidsStringFromModels(plan.Managers)

	response, err := r.client.PostPublisherLicensingLicenseeUpdateWithFormdataBody(ctx, piano_publisher.PostPublisherLicensingLicenseeUpdateFormdataRequestBody{
		Aid:             plan.Aid.ValueString(),
		LicenseeId:      plan.LicenseeId.ValueString(),
		Name:            plan.Name.ValueString(),
		ManagerUids:     managerUidsAsString,
		Representatives: representativesAsString,
		Description:     plan.Description.ValueStringPointer(),
		LogoUrl:         plan.LogoUrl.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update licensee, got error: %s", err))
		tflog.Error(ctx, fmt.Sprintf("Unable to update licensee: %e", err))
		return
	}
	anyResponse, err := syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

	result := piano_publisher.LicenseeResult{}
	err = json.Unmarshal(anyResponse.Raw, &result)
	if err != nil {
		resp.Diagnostics.AddError("Decode Error", fmt.Sprintf("Unable to decode piano AnyMessage into OK Result, got error: %s", err.Error()))
		return
	}

	plan.Name = types.StringValue(result.Licensee.Name)
	plan.Description = types.StringPointerValue(result.Licensee.Description)
	plan.LogoUrl = types.StringPointerValue(result.Licensee.LogoUrl)

	managers := ManagerModelsFromData(result.Licensee.Managers)

	plan.Managers = managers

	representatives := []RepresentativeResourceModel{}
	for _, r := range result.Licensee.Representatives {
		rep := RepresentativeResourceModel{}
		rep.Email = types.StringValue(r.Email)
		representatives = append(representatives, rep)
	}
	plan.Representatives = representatives

	tflog.Info(ctx, fmt.Sprintf("complete updating licensee %s(id: %s)", plan.Name, plan.LicenseeId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LicenseeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LicenseeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("deleting licensee %s:%s in $%s", state.Name.ValueString(), state.LicenseeId.ValueString(), state.Aid.ValueString()))
	response, err := r.client.PostPublisherLicensingLicenseeArchiveWithFormdataBody(ctx, piano_publisher.PostPublisherLicensingLicenseeArchiveFormdataRequestBody{
		Aid:        state.Aid.ValueString(),
		LicenseeId: state.LicenseeId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete licensee, got error: %s", err))
		return
	}
	_, err = syntax.AnyResponseFrom(response, &resp.Diagnostics)
	if err != nil {
		return
	}

}

func (r *LicenseeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceId, err := LicenseeResourceIdFromString(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Licensee resource id", fmt.Sprintf("Unable to parse licensee resource id, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aid"), resourceId.Aid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("licensee_id"), resourceId.LicenseeId)...)
}

func LicenseeManagerUidsStringFromModels(models []ManagerResourceModel) string {
	managerUids := []string{}
	for _, m := range models {
		managerUids = append(managerUids, m.UID.ValueString())
	}
	managerUidsAsString := strings.Join(managerUids, ",")
	return managerUidsAsString
}

func ManagerModelsFromData(items []piano_publisher.LicenseeManager) []ManagerResourceModel {
	managers := []ManagerResourceModel{}
	for _, item := range items {
		manager := ManagerResourceModel{}
		manager.FirstName = types.StringValue(item.FirstName)
		manager.LastName = types.StringValue(item.LastName)
		manager.PersonalName = types.StringValue(item.PersonalName)
		manager.UID = types.StringValue(item.Uid)
		managers = append(managers, manager)
	}
	return managers
}

func LicenseeRepresentativesStringFromModels(models []RepresentativeResourceModel) (*string, error) {
	representatives := []piano_publisher.LicenseeRepresentative{}
	for _, rep := range models {
		representatives = append(representatives, piano_publisher.LicenseeRepresentative{
			Email: rep.Email.ValueString(),
		})
	}
	representativesAsStringBuilder := strings.Builder{}
	err := json.NewEncoder(&representativesAsStringBuilder).Encode(representatives)
	if err != nil {
		return nil, err
	}
	representativesAsString := representativesAsStringBuilder.String()
	return &representativesAsString, nil
}

// LicenseeResourceId represents a piano.io licensee resource identifier in "{aid}/{licensee_id}" format.
type LicenseeResourceId struct {
	Aid        string
	LicenseeId string
}

func LicenseeResourceIdFromString(input string) (*LicenseeResourceId, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, errors.New("Licensee resource id must be in {aid}/{licensee_id} format")
	}
	data := LicenseeResourceId{
		Aid:        parts[0],
		LicenseeId: parts[1],
	}
	return &data, nil
}
