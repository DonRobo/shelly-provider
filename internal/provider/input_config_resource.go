package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jcodybaker/go-shelly"
	"github.com/mongoose-os/mos/common/mgrpc"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &inputConfigResource{}
	_ resource.ResourceWithImportState = &inputConfigResource{}
)

func NewInputConfigResource() resource.Resource {
	return &inputConfigResource{}
}

type inputConfigResourceModel struct {
	IP     types.String `tfsdk:"ip"`
	ID     types.Int32  `tfsdk:"id"`
	Type   types.String `tfsdk:"type"`
	Invert types.Bool   `tfsdk:"invert"`
}

type inputConfigResource struct {
}

func (c *inputConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_input_config"
}

func (c *inputConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The IP address of the Shelly device.",
			},
			"id": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the input to configure.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Type of associated input. Range of values: switch, button, analog, count (only if applicable).",
				Validators: []validator.String{
					stringvalidator.OneOf("switch", "button", "analog", "count"),
				},
			},
			"invert": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "(only for type switch, button, analog) True if the logical state of the associated input is inverted, false otherwise.",
			},
		},
	}
}

func (c *inputConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state inputConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusReq := &shelly.InputGetConfigRequest{
		ID: int(state.ID.ValueInt32()),
	}
	errResult := error(nil)
	WithShellyRPC(ctx, state.IP, &resp.Diagnostics, "InputConfigResource", func(ctxTimeout context.Context, client mgrpc.MgRPC) error {
		statusResp, _, err := statusReq.Do(ctxTimeout, client, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to query device status", err.Error())
			errResult = err
			return err
		}
		state.Type = types.StringValue(*statusResp.Type)
		state.Invert = types.BoolValue(*statusResp.Invert)
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

func setInputConfig(ctx context.Context, plan inputConfigResourceModel, diags *diag.Diagnostics) error {
	var inputConfig shelly.InputConfig
	inputConfig.ID = int(plan.ID.ValueInt32())
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		typeStr := plan.Type.ValueString()
		inputConfig.Type = &typeStr
	}
	if !plan.Invert.IsNull() && !plan.Invert.IsUnknown() {
		invert := plan.Invert.ValueBool()
		inputConfig.Invert = &invert
	}

	enable := true
	inputConfig.Enable = &enable
	statusReq := &shelly.InputSetConfigRequest{
		Config: inputConfig,
	}

	//TODO consider moving the statusReq.Do call into WithShellyRPC
	// to handle the connection and error handling in one place
	errResult := error(nil)
	WithShellyRPC(ctx, plan.IP, diags, "InputConfigResource", func(ctxTimeout context.Context, client mgrpc.MgRPC) error {
		_, _, err := statusReq.Do(ctxTimeout, client, nil)
		if err != nil {
			diags.AddError("Failed to set input config", err.Error())
			errResult = err
			return err
		}
		return nil
	})
	return errResult
}

func (c *inputConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan inputConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := setInputConfig(ctx, plan, &resp.Diagnostics); err != nil {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (c *inputConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan inputConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := setInputConfig(ctx, plan, &resp.Diagnostics); err != nil {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (c *inputConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// TODO test this whole function
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Expected format: ip:id (e.g., 192.168.1.1:123)",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (c *inputConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//TODO We should probably set the input to "no-op" if it's removed, but it's fine for now
	resp.State.RemoveResource(ctx)
}
