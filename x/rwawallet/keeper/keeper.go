package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeService corestoretypes.KVStoreService
	authority    string
}

func NewKeeper(storeService corestoretypes.KVStoreService, authority string) Keeper {
	return Keeper{storeService: storeService, authority: authority}
}

func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Authority() string {
	return k.authority
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	for _, currency := range gs.Currencies {
		if err := k.setCurrency(ctx, currency); err != nil {
			return err
		}
	}
	for _, wallet := range gs.Wallets {
		if err := k.setWallet(ctx, wallet); err != nil {
			return err
		}
	}
	for _, approval := range gs.Approvals {
		if err := k.setApproval(ctx, approval); err != nil {
			return err
		}
	}
	for _, warehouse := range gs.Warehouses {
		if err := k.setWarehouse(ctx, warehouse); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	wallets, err := k.ListWallets(ctx)
	if err != nil {
		return nil, err
	}
	approvals, err := k.ListApprovals(ctx)
	if err != nil {
		return nil, err
	}
	currencies, err := k.ListCurrencies(ctx)
	if err != nil {
		return nil, err
	}
	warehouses, err := k.ListWarehouses(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{
		Authority:  k.authority,
		Wallets:    wallets,
		Approvals:  approvals,
		Currencies: currencies,
		Warehouses: warehouses,
	}
	types.SortData(gs)
	return gs, nil
}

func (k Keeper) SetCurrency(ctx context.Context, currency types.Currency) error {
	if err := currency.Validate(); err != nil {
		return err
	}
	return k.setCurrency(ctx, currency)
}

func (k Keeper) GetCurrencyStatus(ctx context.Context, denom string) (types.Currency, bool, error) {
	return k.getCurrency(ctx, denom)
}

func (k Keeper) QueryApprovals(ctx context.Context) ([]types.Approval, error) {
	return k.ListApprovals(ctx)
}

func (k Keeper) QueryWallet(ctx context.Context, address string) (types.Wallet, bool, error) {
	return k.GetWallet(ctx, address)
}

func (k Keeper) QueryCurrencyStatus(ctx context.Context, denom string) (types.Currency, bool, error) {
	return k.GetCurrencyStatus(ctx, denom)
}

func (k Keeper) QueryTransactionHistory(ctx context.Context, address string) ([]types.TransactionRecord, error) {
	wallet, found, err := k.GetWallet(ctx, address)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrWalletNotFound.Wrap(address)
	}
	return wallet.TransactionHistory, nil
}

func (k Keeper) setWallet(ctx context.Context, wallet types.Wallet) error {
	return k.setJSON(ctx, types.WalletKey(wallet.Address), wallet)
}

func (k Keeper) getWallet(ctx context.Context, address string) (types.Wallet, bool, error) {
	var wallet types.Wallet
	found, err := k.getJSON(ctx, types.WalletKey(address), &wallet)
	return wallet, found, err
}

func (k Keeper) setApproval(ctx context.Context, approval types.Approval) error {
	return k.setJSON(ctx, types.ApprovalKey(approval.WalletAddress), approval)
}

func (k Keeper) getApproval(ctx context.Context, address string) (types.Approval, bool, error) {
	var approval types.Approval
	found, err := k.getJSON(ctx, types.ApprovalKey(address), &approval)
	return approval, found, err
}

func (k Keeper) setCurrency(ctx context.Context, currency types.Currency) error {
	return k.setJSON(ctx, types.CurrencyKey(currency.Denom), currency)
}

func (k Keeper) getCurrency(ctx context.Context, denom string) (types.Currency, bool, error) {
	var currency types.Currency
	found, err := k.getJSON(ctx, types.CurrencyKey(denom), &currency)
	return currency, found, err
}

func (k Keeper) setWarehouse(ctx context.Context, warehouse types.Warehouse) error {
	return k.setJSON(ctx, types.WarehouseKey(warehouse.Name), warehouse)
}

func (k Keeper) getWarehouse(ctx context.Context, name string) (types.Warehouse, bool, error) {
	var warehouse types.Warehouse
	found, err := k.getJSON(ctx, types.WarehouseKey(name), &warehouse)
	return warehouse, found, err
}

func (k Keeper) setJSON(ctx context.Context, key []byte, value any) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return store.Set(key, bz)
}

func (k Keeper) getJSON(ctx context.Context, key []byte, dest any) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(key)
	if err != nil {
		return false, err
	}
	if bz == nil {
		return false, nil
	}
	return true, json.Unmarshal(bz, dest)
}

func listByPrefix[T any](ctx context.Context, service corestoretypes.KVStoreService, prefixBz []byte) ([]T, error) {
	items := make([]T, 0)
	store := prefix.NewStore(runtime.KVStoreAdapter(service.OpenKVStore(ctx)), prefixBz)
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var item T
		if err := json.Unmarshal(iter.Value(), &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func stringInSlice(target string, values []string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
