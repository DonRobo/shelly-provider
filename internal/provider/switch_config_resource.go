// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jcodybaker/go-shelly"
	"github.com/mongoose-os/mos/common/mgrpc"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &switchConfigResource{}
	_ resource.ResourceWithImportState = &switchConfigResource{}
)

func NewSwitchConfigResource() resource.Resource {
	return &switchConfigResource{}
}

type switchConfigResourceModel struct {
	IP           types.String `tfsdk:"ip"`
	ID           types.Int32  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	InMode       types.String `tfsdk:"in_mode"`
	InitialState types.String `tfsdk:"initial_state"`
	//TODO ConsumptionType types.String `tfsdk:"consumption_type"`
}

type switchConfigResource struct {
}

func (c *switchConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_config"
}

func (c *switchConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The IP address of the Shelly device.",
			},
			"id": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "The zero-based ID of the switch to configure (e.g., 0 for the first switch).",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Name of the switch instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"in_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Mode of the associated input",
				Validators: []validator.String{
					stringvalidator.OneOf("momentary", "follow", "flip", "detached", "cycle", "activate"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"initial_state": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Output state to set on power_on",
				Validators: []validator.String{
					stringvalidator.OneOf("off", "on", "restore_last", "match_input"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// TODO
			// "consumption_type": schema.StringAttribute{
			// 	Optional:            true,
			// 	Computed:            true,
			// 	MarkdownDescription: "This setting is mainly used by 3rd party Home Automation systems. Home Assistant supports `light` as an example.",
			// PlanModifiers: []planmodifier.String{
			// 	stringplanmodifier.UseStateForUnknown(),
			// },
			// },
		},
	}
}

func (c *switchConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state switchConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusReq := &shelly.SwitchGetConfigRequest{
		ID: int(state.ID.ValueInt32()),
	}
	errResult := error(nil)
	WithShellyRPC(ctx, state.IP, &resp.Diagnostics, "SwitchConfigResource", func(ctxTimeout context.Context, client mgrpc.MgRPC) error {
		statusResp, _, err := statusReq.Do(ctxTimeout, client, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to query device status", err.Error())
			errResult = err
			return err
		}
		state.Name = types.StringValue(*statusResp.Name)
		state.InMode = types.StringValue(*statusResp.InMode)
		state.InitialState = types.StringValue(*statusResp.InitialState)
		//TODO state.ConsumptionType = types.StringValue(*statusResp.ConsumptionType)
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

func setSwitchConfig(ctx context.Context, plan switchConfigResourceModel, diags *diag.Diagnostics) error {
	var switchConfig shelly.SwitchConfig
	switchConfig.ID = int(plan.ID.ValueInt32())
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		nameStr := plan.Name.ValueString()
		switchConfig.Name = &nameStr
	}
	if !plan.InMode.IsNull() && !plan.InMode.IsUnknown() {
		inModeStr := plan.InMode.ValueString()
		switchConfig.InMode = &inModeStr
	}
	if !plan.InitialState.IsNull() && !plan.InitialState.IsUnknown() {
		initialStateStr := plan.InitialState.ValueString()
		switchConfig.InitialState = &initialStateStr
	}
	//TODO if(!plan.ConsumptionType.IsNull() && !plan.ConsumptionType.IsUnknown()) {
	// 	consumptionTypeStr := plan.ConsumptionType.ValueString()
	// 	switchConfig.ConsumptionType = &consumptionTypeStr
	// }

	statusReq := &shelly.SwitchSetConfigRequest{
		Config: switchConfig,
	}

	//TODO consider moving the statusReq.Do call into WithShellyRPC
	// to handle the connection and error handling in one place
	errResult := error(nil)
	WithShellyRPC(ctx, plan.IP, diags, "SwitchConfigResource", func(ctxTimeout context.Context, client mgrpc.MgRPC) error {
		_, _, err := statusReq.Do(ctxTimeout, client, nil)
		if err != nil {
			diags.AddError("Failed to set switch config", err.Error())
			errResult = err
			return err
		}
		return nil
	})
	return errResult
}

func (c *switchConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan switchConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := setSwitchConfig(ctx, plan, &resp.Diagnostics); err != nil {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (c *switchConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan switchConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := setSwitchConfig(ctx, plan, &resp.Diagnostics); err != nil {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (c *switchConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// TODO test this whole function
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Expected format: ip:id (e.g., 192.168.1.1:123)",
		)
		return
	}

	fmt.Printf("Importing %s. IP=%s, ID=%s\n", req.ID, parts[0], parts[1])

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid switch ID",
			fmt.Sprintf("Could not convert ID '%s' to integer: %v", parts[1], err),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (c *switchConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TODO set in_locked to true and set in_mode to "detached"
	resp.State.RemoveResource(ctx)
}
