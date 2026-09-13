package rwawallet_test

import (
	"testing"

	simapp "github.com/babylonlabs-io/babylon/v4/app"
	"github.com/babylonlabs-io/babylon/v4/x/rwawallet"
	"github.com/stretchr/testify/require"
)

func TestExportGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false)
	genesisState := rwawallet.ExportGenesis(ctx, app.RwaWalletKeeper)
	require.Equal(t, app.RwaWalletKeeper.Authority(), genesisState.Authority)
}
