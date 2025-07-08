// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"terraform-provider-piano/internal/piano_id"
	"terraform-provider-piano/internal/piano_publisher"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure PianoProvider satisfies various provider interfaces.
var _ provider.Provider = &PianoProvider{}
var _ provider.ProviderWithFunctions = &PianoProvider{}
var _ provider.ProviderWithEphemeralResources = &PianoProvider{}

// PianoProvider defines the provider implementation for piano.io resources.
type PianoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PianoProviderModel describes the provider data model.
type PianoProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	ApiToken types.String `tfsdk:"api_token"`
	AppId    types.String `tfsdk:"app_id"`
}

type PianoProviderData struct {
	publisherClient piano_publisher.Client
	idClient        piano_id.Client
}

func (p *PianoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	// The type name should be consistent with resource prefix
	resp.TypeName = "piano"
	resp.Version = p.version
}

func (p *PianoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base endpoint for piano.io API",
				Required:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "API Token for piano.io API",
				Required:            true,
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "App Id for piano.io API",
				Required:            true,
			},
		},
	}
}

func (p *PianoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring piano client")
	var config PianoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown piano API endpoint",
			"The provider cannot create the piano API client as there is an unknown configuration value for the piano API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PIANO_HOST environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := os.Getenv("PIANO_ENDPOINT")
	apiToken := os.Getenv("PIANO_API_TOKEN")
	appId := os.Getenv("PIANO_APP_ID")

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if !config.ApiToken.IsNull() {
		apiToken = config.ApiToken.ValueString()
	}

	tflog.SetField(ctx, "piano_endpoint", endpoint)
	tflog.SetField(ctx, "piano_api_token", apiToken)
	tflog.SetField(ctx, "piano_app_id", appId)
	idEndpoint := fmt.Sprintf("%s/id/api/v1", strings.TrimSuffix(endpoint, "/api/v3"))
	tflog.MaskFieldValuesWithFieldKeys(ctx, "piano_api_token")
	idClient, err := piano_id.NewClient(idEndpoint, func(client *piano_id.Client) error {
		client.RequestEditors = append(client.RequestEditors, func(ctx context.Context, req *http.Request) error {
			copied := req.URL.Query()
			copied.Add("api_token", apiToken)
			copied.Add("aid", appId)
			req.URL.RawQuery = copied.Encode()
			return nil
		})
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Piano id client", fmt.Sprintf("Unable to create Piano id client due to %s", err))
		return
	}
	client, err := piano_publisher.NewClient(endpoint, func(client *piano_publisher.Client) error {
		client.RequestEditors = append(client.RequestEditors, func(ctx context.Context, req *http.Request) error {
			req.Header.Add("API_TOKEN", apiToken)
			return nil
		})
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Piano publisher client", fmt.Sprintf("Unable to create Piano publisher client due to %s", err))
		return
	}
	providerData := PianoProviderData{
		publisherClient: *client,
		idClient:        *idClient,
	}

	resp.ResourceData = providerData
	resp.DataSourceData = providerData
}

func (p *PianoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLicenseeResource,
		NewResourceResource,
		NewContractResource,
		NewPaymentTermResource,
		NewExternalTermResource,
		NewPromotionResource,
		NewOfferResource,
		NewOfferTermBindingResource,
		NewOfferTermOrderResource,
		NewCustomFieldResource,
		NewContractDomainResource,
	}
}

func (p *PianoProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *PianoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLicenseeDataSource,
		NewAppDataSource,
		NewResourceDataSource,
		NewContractDataSource,
		NewTermDataSource,
		NewExternalTermDataSource,
		NewPromotionDataSource,
	}
}

func (p *PianoProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewValidPianoEndpoint,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PianoProvider{
			version: version,
		}
	}
}
