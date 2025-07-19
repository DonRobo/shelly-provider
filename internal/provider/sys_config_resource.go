package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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

	statusReq := &shelly.ShellyGetConfigRequest{}
	errResult := error(nil)
	WithShellyRPC(ctx, state.IP, &resp.Diagnostics, "ConfigResource", func(ctxTimeout context.Context, client mgrpc.MgRPC) error {
		statusResp, _, err := statusReq.Do(ctxTimeout, client, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to query device status", err.Error())
			errResult = err
			return err
		}
		if statusResp.System.Device.Name == nil {
			state.Name = types.StringNull()
		} else {
			state.Name = types.StringValue(*statusResp.System.Device.Name)
		}
		return nil
	})
	if errResult != nil {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func setDeviceConfig(ctx context.Context, plan configResourceModel, diags *diag.Diagnostics) error {
	var deviceConfig shelly.SysDeviceConfig
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		nameStr := plan.Name.ValueString()
		deviceConfig.Name = &nameStr
	}
	// Add more fields here as you extend configResourceModel

	statusReq := &shelly.SysSetConfigRequest{
		Config: shelly.SysConfig{
			Device: &deviceConfig,
		},
	}

	errResult := error(nil)
	WithShellyRPC(ctx, plan.IP, diags, "ConfigResource", func(ctxTimeout context.Context, client mgrpc.MgRPC) error {
		_, _, err := statusReq.Do(ctxTimeout, client, nil)
		if err != nil {
			diags.AddError("Failed to set device config", err.Error())
			errResult = err
			return err
		}
		return nil
	})
	return errResult
}

func (c *configResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := setDeviceConfig(ctx, plan, &resp.Diagnostics); err != nil {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (c *configResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := setDeviceConfig(ctx, plan, &resp.Diagnostics); err != nil {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (c *configResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("ip"), req, resp)
}

func (c *configResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}
