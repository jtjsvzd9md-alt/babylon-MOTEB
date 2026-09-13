package rwawallet

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/gorilla/mux"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/keeper"
	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
	_ module.HasABCIEndBlock    = AppModule{}
	_ module.AppModuleBasic     = AppModuleBasic{}
)

type AppModuleBasic struct {
	k keeper.Keeper
}

func NewAppModuleBasic(k keeper.Keeper) AppModuleBasic {
	return AppModuleBasic{k: k}
}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {}

func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(types.DefaultGenesis(a.k.Authority()))
	if err != nil {
		panic(err)
	}
	return bz
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := json.Unmarshal(bz, &genState); err != nil {
		return err
	}
	return genState.Validate()
}

func (AppModuleBasic) RegisterRESTRoutes(client.Context, *mux.Router) {}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *gwruntime.ServeMux) {}

func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(_ codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{AppModuleBasic: NewAppModuleBasic(keeper), keeper: keeper}
}

func (am AppModule) Name() string { return am.AppModuleBasic.Name() }

func (AppModule) QuerierRoute() string { return types.QuerierRoute }

func (AppModule) RegisterServices(cfg module.Configurator) {}

func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	if err := json.Unmarshal(gs, &genState); err != nil {
		panic(err)
	}
	InitGenesis(ctx, am.keeper, genState)
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(ExportGenesis(ctx, am.keeper))
	if err != nil {
		panic(err)
	}
	return bz
}

func (AppModule) ConsensusVersion() uint64 { return 1 }

func (AppModule) BeginBlock(_ context.Context) error { return nil }

func (AppModule) EndBlock(_ context.Context) ([]abci.ValidatorUpdate, error) {
	return []abci.ValidatorUpdate{}, nil
}

func (AppModule) IsOnePerModuleType() {}

func (AppModule) IsAppModule() {}
