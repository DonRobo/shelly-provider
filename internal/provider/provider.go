// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ShellyProvider{}

// ShellyProvider defines the provider implementation.
type ShellyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ShellyProviderModel describes the provider data model.
type ShellyProviderModel struct {
	IP types.String `tfsdk:"ip"`
}

func (p *ShellyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "shelly"
	resp.Version = p.version
}

func (p *ShellyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "IP address of the Shelly.",
				Required:            true,
			},
		},
	}
}

func (p *ShellyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ShellyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...) // get config

	if resp.Diagnostics.HasError() {
		return
	}

	if data.IP.IsNull() || data.IP.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing IP address",
			"The provider requires an 'ip' attribute specifying the Shelly Gen2 device IP.",
		)
		return
	}

	resp.DataSourceData = map[string]any{
		"ip": data.IP.ValueString(),
	}
	resp.ResourceData = resp.DataSourceData
}

func (p *ShellyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *ShellyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewShellyVersionDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ShellyProvider{
			version: version,
		}
	}
}
