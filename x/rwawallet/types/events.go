package types

const (
	EventTypeWalletCreated   = "rwa_wallet_created"
	EventTypeWalletApproved  = "rwa_wallet_approved"
	EventTypeApprovalRevoked = "rwa_wallet_revoked"
	EventTypeTransferred     = "rwa_wallet_transferred"
	EventTypeCurrencyRotated = "rwa_currency_rotated"
	EventTypeWarehouseLinked = "rwa_warehouse_linked"

	AttributeKeyWallet    = "wallet"
	AttributeKeyOwner     = "owner"
	AttributeKeyAuthority = "authority"
	AttributeKeyWarehouse = "warehouse"
	AttributeKeyDenom     = "denom"
	AttributeKeyState     = "state"
	AttributeKeyCycle     = "cycle"
)
