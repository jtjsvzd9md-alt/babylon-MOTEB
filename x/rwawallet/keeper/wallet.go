package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) CreateWallet(ctx context.Context, msg types.MsgCreateWallet) (types.Wallet, error) {
	if err := msg.Validate(); err != nil {
		return types.Wallet{}, err
	}
	if _, found, err := k.getWallet(ctx, msg.Address); err != nil {
		return types.Wallet{}, err
	} else if found {
		return types.Wallet{}, types.ErrWalletExists.Wrap(msg.Address)
	}
	createdAt := msg.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	wallet := types.Wallet{
		Address:            msg.Address,
		Owner:              msg.Owner,
		Warehouse:          msg.Warehouse,
		Balances:           msg.InitialBalances.Sort(),
		TransactionHistory: []types.TransactionRecord{},
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
	}
	if err := wallet.Validate(); err != nil {
		return types.Wallet{}, err
	}
	if err := k.setWallet(ctx, wallet); err != nil {
		return types.Wallet{}, err
	}
	approval := types.Approval{WalletAddress: wallet.Address, RequestedBy: wallet.Owner, RequestedAt: createdAt}
	if err := k.setApproval(ctx, approval); err != nil {
		return types.Wallet{}, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeWalletCreated,
		sdk.NewAttribute(types.AttributeKeyWallet, wallet.Address),
		sdk.NewAttribute(types.AttributeKeyOwner, wallet.Owner),
		sdk.NewAttribute(types.AttributeKeyWarehouse, wallet.Warehouse),
	))
	return wallet, nil
}

func (k Keeper) GetWallet(ctx context.Context, address string) (types.Wallet, bool, error) {
	return k.getWallet(ctx, address)
}

func (k Keeper) ListWallets(ctx context.Context) ([]types.Wallet, error) {
	return listByPrefix[types.Wallet](ctx, k.storeService, types.WalletPrefix)
}

func (k Keeper) InitializeWarehouseWallet(ctx context.Context, warehouseName, walletAddress, owner string, supportedDenoms []string, initializedAt time.Time) (types.Warehouse, error) {
	wallet, found, err := k.GetWallet(ctx, walletAddress)
	if err != nil {
		return types.Warehouse{}, err
	}
	if !found {
		wallet, err = k.CreateWallet(ctx, types.MsgCreateWallet{
			Address:         walletAddress,
			Owner:           owner,
			Warehouse:       warehouseName,
			InitialBalances: sdk.NewCoins(),
			CreatedAt:       initializedAt,
		})
		if err != nil {
			return types.Warehouse{}, err
		}
	} else if wallet.Warehouse != "" && wallet.Warehouse != warehouseName {
		return types.Warehouse{}, types.ErrWarehouseWalletMismatch.Wrapf("wallet %s already linked to %s", walletAddress, wallet.Warehouse)
	}
	if initializedAt.IsZero() {
		initializedAt = time.Now().UTC()
	}
	wallet.Warehouse = warehouseName
	wallet.UpdatedAt = initializedAt
	if err := k.setWallet(ctx, wallet); err != nil {
		return types.Warehouse{}, err
	}
	warehouse := types.Warehouse{
		Name:            warehouseName,
		WalletAddress:   walletAddress,
		SupportedDenoms: supportedDenoms,
		InitializedAt:   initializedAt,
	}
	if err := warehouse.Validate(); err != nil {
		return types.Warehouse{}, err
	}
	if err := k.setWarehouse(ctx, warehouse); err != nil {
		return types.Warehouse{}, err
	}
	for _, denom := range supportedDenoms {
		currency, found, err := k.getCurrency(ctx, denom)
		if err != nil {
			return types.Warehouse{}, err
		}
		if found && !stringInSlice(warehouseName, currency.SupportedWarehouses) {
			currency.SupportedWarehouses = append(currency.SupportedWarehouses, warehouseName)
			if err := k.setCurrency(ctx, currency); err != nil {
				return types.Warehouse{}, err
			}
		}
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeWarehouseLinked,
		sdk.NewAttribute(types.AttributeKeyWarehouse, warehouseName),
		sdk.NewAttribute(types.AttributeKeyWallet, walletAddress),
	))
	return warehouse, nil
}

func (k Keeper) GetWarehouse(ctx context.Context, name string) (types.Warehouse, bool, error) {
	return k.getWarehouse(ctx, name)
}

func (k Keeper) ListWarehouses(ctx context.Context) ([]types.Warehouse, error) {
	return listByPrefix[types.Warehouse](ctx, k.storeService, types.WarehousePrefix)
}

func (k Keeper) ListCurrencies(ctx context.Context) ([]types.Currency, error) {
	return listByPrefix[types.Currency](ctx, k.storeService, types.CurrencyPrefix)
}

func (k Keeper) Transfer(ctx context.Context, msg types.MsgTransfer) (types.TransactionRecord, error) {
	if err := msg.Validate(); err != nil {
		return types.TransactionRecord{}, err
	}
	fromWallet, found, err := k.GetWallet(ctx, msg.FromAddress)
	if err != nil {
		return types.TransactionRecord{}, err
	}
	if !found {
		return types.TransactionRecord{}, types.ErrWalletNotFound.Wrap(msg.FromAddress)
	}
	toWallet, found, err := k.GetWallet(ctx, msg.ToAddress)
	if err != nil {
		return types.TransactionRecord{}, err
	}
	if !found {
		return types.TransactionRecord{}, types.ErrWalletNotFound.Wrap(msg.ToAddress)
	}
	approval, found, err := k.GetApproval(ctx, msg.FromAddress)
	if err != nil {
		return types.TransactionRecord{}, err
	}
	if !found || !approval.Approved {
		return types.TransactionRecord{}, types.ErrApprovalRequired.Wrap(msg.FromAddress)
	}
	currency, found, err := k.getCurrency(ctx, msg.Amount.Denom)
	if err != nil {
		return types.TransactionRecord{}, err
	}
	if !found {
		return types.TransactionRecord{}, types.ErrCurrencyNotFound.Wrap(msg.Amount.Denom)
	}
	if currency.State != types.CurrencyStateActive {
		return types.TransactionRecord{}, types.ErrCurrencyInactive.Wrapf("currency %s is %s", currency.Denom, currency.State)
	}
	if err := k.ensureWarehouseCompatibility(ctx, fromWallet, msg.Amount.Denom); err != nil {
		return types.TransactionRecord{}, err
	}
	if err := k.ensureWarehouseCompatibility(ctx, toWallet, msg.Amount.Denom); err != nil {
		return types.TransactionRecord{}, err
	}
	updatedBalances, hasNeg := fromWallet.Balances.SafeSub(msg.Amount)
	if hasNeg {
		return types.TransactionRecord{}, types.ErrInsufficientBalance.Wrapf("wallet %s balance %s", fromWallet.Address, fromWallet.Balances.String())
	}
	fromWallet.Balances = updatedBalances.Sort()
	toWallet.Balances = toWallet.Balances.Add(msg.Amount).Sort()
	executedAt := msg.ExecutedAt
	if executedAt.IsZero() {
		executedAt = time.Now().UTC()
	}
	txRecord := types.TransactionRecord{
		ID:         fmt.Sprintf("%s-%d", msg.Executor, executedAt.UnixNano()),
		From:       fromWallet.Address,
		To:         toWallet.Address,
		Amount:     msg.Amount,
		Cycle:      currency.Cycle,
		State:      currency.State,
		ApprovedBy: approval.ApprovedBy,
		Executor:   msg.Executor,
		Timestamp:  executedAt,
	}
	fromWallet.TransactionHistory = append(fromWallet.TransactionHistory, txRecord)
	toWallet.TransactionHistory = append(toWallet.TransactionHistory, txRecord)
	fromWallet.UpdatedAt = executedAt
	toWallet.UpdatedAt = executedAt
	if err := k.setWallet(ctx, fromWallet); err != nil {
		return types.TransactionRecord{}, err
	}
	if err := k.setWallet(ctx, toWallet); err != nil {
		return types.TransactionRecord{}, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransferred,
		sdk.NewAttribute(sdk.AttributeKeySender, fromWallet.Address),
		sdk.NewAttribute("recipient", toWallet.Address),
		sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyCycle, fmt.Sprintf("%d", currency.Cycle)),
	))
	return txRecord, nil
}

func (k Keeper) RotateCurrency(ctx context.Context, authority, denom string, rotatedAt time.Time) (types.Currency, error) {
	if authority != k.authority {
		return types.Currency{}, types.ErrUnauthorized.Wrapf("expected %s", k.authority)
	}
	currency, found, err := k.getCurrency(ctx, denom)
	if err != nil {
		return types.Currency{}, err
	}
	if !found {
		return types.Currency{}, types.ErrCurrencyNotFound.Wrap(denom)
	}
	if rotatedAt.IsZero() {
		rotatedAt = time.Now().UTC()
	}
	if currency.CycleDuration > 0 && !currency.LastRotatedAt.IsZero() && rotatedAt.Sub(currency.LastRotatedAt) < currency.CycleDuration {
		return types.Currency{}, types.ErrRotationTooEarly.Wrapf("remaining %s", currency.CycleDuration-rotatedAt.Sub(currency.LastRotatedAt))
	}
	nextState, err := types.NextCurrencyState(currency.State)
	if err != nil {
		return types.Currency{}, err
	}
	currency.State = nextState
	currency.Cycle++
	currency.LastRotatedAt = rotatedAt
	if err := k.setCurrency(ctx, currency); err != nil {
		return types.Currency{}, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCurrencyRotated,
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyState, currency.State),
		sdk.NewAttribute(types.AttributeKeyCycle, fmt.Sprintf("%d", currency.Cycle)),
	))
	return currency, nil
}

func (k Keeper) ensureWarehouseCompatibility(ctx context.Context, wallet types.Wallet, denom string) error {
	if wallet.Warehouse == "" {
		return nil
	}
	warehouse, found, err := k.GetWarehouse(ctx, wallet.Warehouse)
	if err != nil {
		return err
	}
	if !found {
		return types.ErrWarehouseNotFound.Wrap(wallet.Warehouse)
	}
	if !stringInSlice(denom, warehouse.SupportedDenoms) {
		return types.ErrCurrencyIncompatible.Wrapf("warehouse %s does not support %s", warehouse.Name, denom)
	}
	return nil
}
