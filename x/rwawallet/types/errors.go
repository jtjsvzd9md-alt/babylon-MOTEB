package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrWalletExists            = errorsmod.Register(ModuleName, 1, "wallet already exists")
	ErrWalletNotFound          = errorsmod.Register(ModuleName, 2, "wallet not found")
	ErrUnauthorized            = errorsmod.Register(ModuleName, 3, "unauthorized")
	ErrApprovalRequired        = errorsmod.Register(ModuleName, 4, "wallet approval required")
	ErrCurrencyNotFound        = errorsmod.Register(ModuleName, 5, "currency not found")
	ErrCurrencyInactive        = errorsmod.Register(ModuleName, 6, "currency is not active")
	ErrInsufficientBalance     = errorsmod.Register(ModuleName, 7, "insufficient balance")
	ErrWarehouseNotFound       = errorsmod.Register(ModuleName, 8, "warehouse not found")
	ErrCurrencyIncompatible    = errorsmod.Register(ModuleName, 9, "currency not supported by warehouse")
	ErrRotationTooEarly        = errorsmod.Register(ModuleName, 10, "currency rotation attempted before cycle duration elapsed")
	ErrInvalidCurrencyState    = errorsmod.Register(ModuleName, 11, "invalid currency state")
	ErrApprovalRecordNotFound  = errorsmod.Register(ModuleName, 12, "approval record not found")
	ErrWarehouseWalletMismatch = errorsmod.Register(ModuleName, 13, "warehouse wallet mismatch")
	ErrDuplicateGenesisEntry   = errorsmod.Register(ModuleName, 14, "duplicate genesis entry")
)
