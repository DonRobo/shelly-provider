package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ShellyVersionDataSource struct {
	ip     string
	client *http.Client
}

type ShellyVersionModel struct {
	Version types.String `tfsdk:"version"`
}

func NewShellyVersionDataSource() datasource.DataSource {
	return &ShellyVersionDataSource{}
}

func (d *ShellyVersionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "shelly_version"
}

func (d *ShellyVersionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The firmware version of the Shelly Gen2 device.",
			},
		},
	}
}

func (d *ShellyVersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("Provider data error", "Could not get provider data.")
		return
	}
	ip, ok := providerData["ip"].(string)
	if !ok || ip == "" {
		resp.Diagnostics.AddError("Missing IP", "Provider did not supply a valid IP address.")
		return
	}
	client, ok := providerData["client"].(*http.Client)
	if !ok {
		resp.Diagnostics.AddError("HTTP client error", "Could not get HTTP client from provider data.")
		return
	}
	d.ip = ip
	d.client = client
}

func (d *ShellyVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	data := &ShellyVersionModel{}

	if d.ip == "" || d.client == nil {
		resp.Diagnostics.AddError("Not configured", "The data source is not configured with provider data.")
		return
	}

	url := fmt.Sprintf("http://%s/rpc/Shelly.GetStatus", d.ip)
	respHTTP, err := d.client.Get(url)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request failed", err.Error())
		return
	}
	defer respHTTP.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(respHTTP.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	version, ok := result["fw_id"].(string)
	if !ok {
		resp.Diagnostics.AddError("Version not found", "Could not find 'fw_id' in response.")
		return
	}
	data.Version = types.StringValue(version)
	resp.State.Set(ctx, data)
}
