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

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ShellyDeviceDataSource{}

type ShellyDeviceDataSource struct {
}

type ShellyDeviceModel struct {
	IP      types.String `tfsdk:"ip"`
	MAC     types.String `tfsdk:"mac"`
	Version types.String `tfsdk:"version"`
}

func NewShellyDeviceDataSource() datasource.DataSource {
	return &ShellyDeviceDataSource{}
}

func (d *ShellyDeviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "shelly_device"
}

func (d *ShellyDeviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The IP address of the device.",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The firmware version the device.",
			},
			"mac": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MAC address of the device.",
			},
		},
	}
}

func (d *ShellyDeviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
}

func (d *ShellyDeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	data := &ShellyDeviceModel{}

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Do RPC call
	statusReq := &shelly.SysGetConfigRequest{}
	rpcAddr := fmt.Sprintf("http://%s/rpc", data.IP.ValueString())
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	fmt.Printf("[ShellyDeviceDataSource] Creating mgrpc client for %s\n", rpcAddr)
	c, err := mgrpc.New(ctxTimeout, rpcAddr, mgrpc.UseHTTPPost())
	if err != nil {
		resp.Diagnostics.AddError("Failed to establish RPC channel", err.Error())
		fmt.Printf("[ShellyDeviceDataSource] RPC client error: %v\n", err)
		return
	}
	defer c.Disconnect(ctxTimeout)
	fmt.Printf("[ShellyDeviceDataSource] Making RPC call with client: %v\n", c)
	statusResp, _, err := statusReq.Do(ctxTimeout, c, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to query device status", err.Error())
		fmt.Printf("[ShellyDeviceDataSource] RPC error: %v\n", err)
		return
	}

	data.Version = types.StringValue(statusResp.Device.FW_ID)
	if data.Version.IsNull() || data.Version.IsUnknown() || data.Version.ValueString() == "" {
		resp.Diagnostics.AddError("Version not found", "Could not find valid firmware version in response.")
		return
	}

	data.MAC = types.StringValue(statusResp.Device.Mac)
	if data.MAC.IsNull() || data.MAC.IsUnknown() || data.MAC.ValueString() == "" {
		resp.Diagnostics.AddError("MAC address not found", "Could not find valid MAC address in response.")
		return
	}

	// Write to state
	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
