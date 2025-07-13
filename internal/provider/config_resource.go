package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jcodybaker/go-shelly"
	"github.com/mongoose-os/mos/common/mgrpc"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &configResource{}
	_ resource.ResourceWithImportState = &configResource{}
)

func NewConfigResource() resource.Resource {
	return &configResource{}
}

type configResourceModel struct {
	IP   types.String `tfsdk:"ip"`
	Name types.String `tfsdk:"name"`
}

type configResource struct {
}

func (c *configResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (c *configResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The IP address of the Shelly device.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name of the Shelly device.",
			},
		},
	}
}

func (c *configResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Do RPC call
	statusReq := &shelly.ShellyGetConfigRequest{}
	rpcAddr := fmt.Sprintf("http://%s/rpc", state.IP.ValueString())
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	fmt.Printf("[ShellyDeviceDataSource] Creating mgrpc client for %s\n", rpcAddr)
	con, err := mgrpc.New(ctxTimeout, rpcAddr, mgrpc.UseHTTPPost())
	if err != nil {
		resp.Diagnostics.AddError("Failed to establish RPC channel", err.Error())
		fmt.Printf("[ShellyDeviceDataSource] RPC client error: %v\n", err)
		return
	}
	defer con.Disconnect(ctxTimeout)
	fmt.Printf("[ShellyDeviceDataSource] Making RPC call with client: %v\n", con)
	statusResp, _, err := statusReq.Do(ctxTimeout, con, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to query device status", err.Error())
		fmt.Printf("[ShellyDeviceDataSource] RPC error: %v\n", err)
		return
	}

	if statusResp.System.Device.Name == nil {
		state.Name = types.StringNull()
	} else {
		state.Name = types.StringValue(*statusResp.System.Device.Name)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (c *configResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Do RPC call
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		resp.Diagnostics.AddError("Invalid Name", "The 'name' attribute must be set to a valid string.")
		return
	}
	name := plan.Name.ValueString()
	statusReq := &shelly.SysSetConfigRequest{
		Config: shelly.SysConfig{
			Device: &shelly.SysDeviceConfig{
				Name: &name,
			},
		},
	}
	rpcAddr := fmt.Sprintf("http://%s/rpc", plan.IP.ValueString())
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	fmt.Printf("[ShellyDeviceDataSource] Creating mgrpc client for %s\n", rpcAddr)
	con, err := mgrpc.New(ctxTimeout, rpcAddr, mgrpc.UseHTTPPost())
	if err != nil {
		resp.Diagnostics.AddError("Failed to establish RPC channel", err.Error())
		fmt.Printf("[ShellyDeviceDataSource] RPC client error: %v\n", err)
		return
	}
	defer con.Disconnect(ctxTimeout)
	fmt.Printf("[ShellyDeviceDataSource] Making RPC call with client: %v\n", con)
	_, _, err = statusReq.Do(ctxTimeout, con, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to query device status", err.Error())
		fmt.Printf("[ShellyDeviceDataSource] RPC error: %v\n", err)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (c *configResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("ip"), req, resp)
}

func (c *configResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
	panic("unimplemented")
}

func (c *configResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	panic("unimplemented")
}
