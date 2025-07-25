package provider

import (
	"context"
	"fmt"
	"time"

	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mongoose-os/mos/common/mgrpc"
)

// Helper for Shelly RPC client creation and error handling.
func WithShellyRPC(ctx context.Context, ip types.String, diags *diag.Diagnostics, logPrefix string, rpcFunc func(ctxTimeout context.Context, client mgrpc.MgRPC) error) {
	rpcAddr := fmt.Sprintf("http://%s/rpc", ip.ValueString())
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	fmt.Printf("[%s] Creating mgrpc client for %s\n", logPrefix, rpcAddr)
	client, err := mgrpc.New(ctxTimeout, rpcAddr, mgrpc.UseHTTPPost())
	if err != nil {
		diags.AddError("Failed to establish RPC channel", err.Error())
		fmt.Printf("[%s] RPC client error: %v\n", logPrefix, err)
		return
	}
	defer func() {
		if err := client.Disconnect(ctxTimeout); err != nil {
			fmt.Printf("[%s] Error disconnecting RPC client: %v\n", logPrefix, err)
		}
	}()
	fmt.Printf("[%s] Making RPC call with client: %v\n", logPrefix, client)
	if err := rpcFunc(ctxTimeout, client); err != nil {
		fmt.Printf("[%s] RPC error: %v\n", logPrefix, err)
	}
}
