package keeper_test

import (
	"testing"
	"time"

	keepertest "github.com/babylonlabs-io/babylon/v4/testutil/keeper"
	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestWalletApprovalAndTransferLifecycle(t *testing.T) {
	const authority = "babylon1authority"
	k, ctx := keepertest.RwaWalletKeeper(t, authority)
	now := time.Now().UTC().Truncate(time.Second)

	err := k.SetCurrency(ctx, types.Currency{
		Denom:               "urwa",
		State:               types.CurrencyStateActive,
		Cycle:               1,
		CycleDuration:       time.Hour,
		LastRotatedAt:       now.Add(-2 * time.Hour),
		SupportedWarehouses: []string{"warehouse-a", "warehouse-b"},
	})
	require.NoError(t, err)

	_, err = k.InitializeWarehouseWallet(ctx, "warehouse-a", "wallet-a", "owner-a", []string{"urwa"}, now)
	require.NoError(t, err)
	_, err = k.InitializeWarehouseWallet(ctx, "warehouse-b", "wallet-b", "owner-b", []string{"urwa"}, now)
	require.NoError(t, err)

	walletA, err := k.CreateWallet(ctx, types.MsgCreateWallet{
		Address:         "wallet-user-a",
		Owner:           "owner-a",
		Warehouse:       "warehouse-a",
		InitialBalances: sdk.NewCoins(sdk.NewInt64Coin("urwa", 100)),
		CreatedAt:       now,
	})
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("urwa", 100)), walletA.Balances)

	_, err = k.CreateWallet(ctx, types.MsgCreateWallet{
		Address:         "wallet-user-b",
		Owner:           "owner-b",
		Warehouse:       "warehouse-b",
		InitialBalances: sdk.NewCoins(),
		CreatedAt:       now,
	})
	require.NoError(t, err)

	_, err = k.Transfer(ctx, types.MsgTransfer{
		FromAddress: "wallet-user-a",
		ToAddress:   "wallet-user-b",
		Amount:      sdk.NewInt64Coin("urwa", 5),
		Executor:    "owner-a",
		ExecutedAt:  now,
	})
	require.ErrorIs(t, err, types.ErrApprovalRequired)

	approval, err := k.ApproveWallet(ctx, authority, "wallet-user-a", now)
	require.NoError(t, err)
	require.True(t, approval.Approved)

	_, err = k.Transfer(ctx, types.MsgTransfer{
		FromAddress: "wallet-user-a",
		ToAddress:   "wallet-user-b",
		Amount:      sdk.NewInt64Coin("urwa", 5),
		Executor:    "owner-a",
		ExecutedAt:  now.Add(time.Minute),
	})
	require.ErrorIs(t, err, types.ErrApprovalRequired)

	_, err = k.ApproveWallet(ctx, authority, "wallet-user-b", now)
	require.NoError(t, err)

	txRecord, err := k.Transfer(ctx, types.MsgTransfer{
		FromAddress: "wallet-user-a",
		ToAddress:   "wallet-user-b",
		Amount:      sdk.NewInt64Coin("urwa", 5),
		Executor:    "owner-a",
		ExecutedAt:  now.Add(2 * time.Minute),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), txRecord.Cycle)
	require.Equal(t, authority, txRecord.ApprovedBy)

	fromWallet, found, err := k.GetWallet(ctx, "wallet-user-a")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("urwa", 95)), fromWallet.Balances)
	require.Len(t, fromWallet.TransactionHistory, 1)

	toWallet, found, err := k.GetWallet(ctx, "wallet-user-b")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("urwa", 5)), toWallet.Balances)
	require.Len(t, toWallet.TransactionHistory, 1)

	revoked, err := k.RevokeApproval(ctx, authority, "wallet-user-a", now.Add(2*time.Minute))
	require.NoError(t, err)
	require.False(t, revoked.Approved)
}

func TestCurrencyRotationAndGenesisExport(t *testing.T) {
	const authority = "babylon1authority"
	k, ctx := keepertest.RwaWalletKeeper(t, authority)
	now := time.Now().UTC().Truncate(time.Second)

	err := k.InitGenesis(ctx, types.GenesisState{
		Authority: authority,
		Wallets: []types.Wallet{{
			Address:   "warehouse-wallet-a",
			Owner:     "owner-a",
			Warehouse: "warehouse-a",
		}},
		Currencies: []types.Currency{{
			Denom:               "urwa",
			State:               types.CurrencyStateActive,
			Cycle:               2,
			CycleDuration:       time.Hour,
			LastRotatedAt:       now.Add(-2 * time.Hour),
			SupportedWarehouses: []string{"warehouse-a"},
		}},
		Warehouses: []types.Warehouse{{
			Name:            "warehouse-a",
			WalletAddress:   "warehouse-wallet-a",
			SupportedDenoms: []string{"urwa"},
		}},
	})
	require.NoError(t, err)

	currency, err := k.RotateCurrency(ctx, authority, "urwa", now)
	require.NoError(t, err)
	require.Equal(t, types.CurrencyStateInactive, currency.State)
	require.Equal(t, uint64(3), currency.Cycle)

	_, err = k.RotateCurrency(ctx, authority, "urwa", now.Add(30*time.Minute))
	require.ErrorIs(t, err, types.ErrRotationTooEarly)

	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, authority, exported.Authority)
	require.Len(t, exported.Currencies, 1)
	require.Equal(t, types.CurrencyStateInactive, exported.Currencies[0].State)
}

func TestWarehouseCompatibilityBlocksUnsupportedTransfers(t *testing.T) {
	const authority = "babylon1authority"
	k, ctx := keepertest.RwaWalletKeeper(t, authority)
	now := time.Now().UTC()

	require.NoError(t, k.SetCurrency(ctx, types.Currency{Denom: "urwa", State: types.CurrencyStateActive}))
	_, err := k.InitializeWarehouseWallet(ctx, "warehouse-a", "wallet-a", "owner-a", []string{"urwa"}, now)
	require.NoError(t, err)
	_, err = k.InitializeWarehouseWallet(ctx, "warehouse-b", "wallet-b", "owner-b", []string{"uother"}, now)
	require.NoError(t, err)
	_, err = k.CreateWallet(ctx, types.MsgCreateWallet{Address: "user-a", Owner: "owner-a", Warehouse: "warehouse-a", InitialBalances: sdk.NewCoins(sdk.NewInt64Coin("urwa", 10)), CreatedAt: now})
	require.NoError(t, err)
	_, err = k.CreateWallet(ctx, types.MsgCreateWallet{Address: "user-b", Owner: "owner-b", Warehouse: "warehouse-b", InitialBalances: sdk.NewCoins(), CreatedAt: now})
	require.NoError(t, err)
	_, err = k.ApproveWallet(ctx, authority, "user-a", now)
	require.NoError(t, err)

	_, err = k.Transfer(ctx, types.MsgTransfer{FromAddress: "user-a", ToAddress: "user-b", Amount: sdk.NewInt64Coin("urwa", 1), Executor: "owner-a", ExecutedAt: now})
	require.ErrorIs(t, err, types.ErrApprovalRequired)

	_, err = k.ApproveWallet(ctx, authority, "user-b", now)
	require.NoError(t, err)

	_, err = k.Transfer(ctx, types.MsgTransfer{FromAddress: "user-a", ToAddress: "user-b", Amount: sdk.NewInt64Coin("urwa", 1), Executor: "owner-a", ExecutedAt: now})
	require.ErrorIs(t, err, types.ErrCurrencyIncompatible)
}

func TestInitGenesisRejectsBrokenReferences(t *testing.T) {
	const authority = "babylon1authority"
	k, ctx := keepertest.RwaWalletKeeper(t, authority)

	err := k.InitGenesis(ctx, types.GenesisState{
		Authority: authority,
		Wallets: []types.Wallet{{
			Address:   "wallet-user-a",
			Owner:     "owner-a",
			Warehouse: "warehouse-a",
		}},
		Approvals: []types.Approval{{
			WalletAddress: "missing-wallet",
			RequestedBy:   "owner-a",
		}},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "wallet")
}
