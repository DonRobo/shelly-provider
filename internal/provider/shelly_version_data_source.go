package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jcodybaker/go-shelly"
	"github.com/mongoose-os/mos/common/mgrpc"
)

type ShellyVersionDataSource struct {
	ip string
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
	d.ip = ip
}

func (d *ShellyVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	data := &ShellyVersionModel{}

	if d.ip == "" {
		resp.Diagnostics.AddError("Not configured", "The data source is not configured with provider data.")
		return
	}

	statusReq := &shelly.ShellyGetConfigRequest{}
	rpcAddr := fmt.Sprintf("http://%s/rpc", d.ip)
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	fmt.Printf("[ShellyVersionDataSource] Creating mgrpc client for %s\n", rpcAddr)
	c, err := mgrpc.New(ctxTimeout, rpcAddr, mgrpc.UseHTTPPost())
	if err != nil {
		resp.Diagnostics.AddError("Failed to establish RPC channel", err.Error())
		fmt.Printf("[ShellyVersionDataSource] RPC client error: %v\n", err)
		return
	}
	defer c.Disconnect(ctxTimeout)
	fmt.Printf("[ShellyVersionDataSource] Making RPC call with client: %v\n", c)
	statusResp, _, err := statusReq.Do(ctxTimeout, c, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to query device status", err.Error())
		fmt.Printf("[ShellyVersionDataSource] RPC error: %v\n", err)
		return
	}

	version := statusResp.System.Device.FW_ID
	if version == "" {
		resp.Diagnostics.AddError("Version not found", "Could not find 'fw_id' in response.")
		return
	}
	data.Version = types.StringValue(version)
	resp.State.Set(ctx, data)
}
