package rwawallet

import (
	"context"

	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/keeper"
	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/types"
)

func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}
	if err := k.InitGenesis(ctx, genState); err != nil {
		panic(err)
	}
}

func ExportGenesis(ctx context.Context, k keeper.Keeper) *types.GenesisState {
	gs, err := k.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return gs
}
