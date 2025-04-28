// Copyright (c) Yoichiro Ito <contact.110416@gmail.com>
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
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
}

func (p *PianoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	// The type name should be consistent with resource prefix
	resp.TypeName = "piano"
	resp.Version = p.version
}

func (p *PianoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	// ThHis value is `true` for debug purpose
	required := true
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base endpoint for piano.io API",
				Required:            required,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "API Token for piano.io API",
				Required:            required,
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

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if !config.ApiToken.IsNull() {
		apiToken = config.ApiToken.ValueString()
	}

	tflog.SetField(ctx, "piano_endpoint", endpoint)
	tflog.SetField(ctx, "piano_api_token", apiToken)
	tflog.MaskFieldValuesWithFieldKeys(ctx, "piano_api_token")
	client, err := piano_publisher.NewClient(endpoint, func(client *piano_publisher.Client) error {
		client.RequestEditors = append(client.RequestEditors, func(ctx context.Context, req *http.Request) error {
			req.Header.Add("API_TOKEN", apiToken)
			return nil
		})
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Piano client", fmt.Sprintf("Unable to create Piano client due to %s", err))
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PianoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLicenseeResource,
		NewResourceResource,
		NewContractResource,
		NewPaymentTermResource,
		NewExternalTermResource,
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
