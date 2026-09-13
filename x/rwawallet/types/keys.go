package types

var (
	WalletPrefix    = []byte{1}
	ApprovalPrefix  = []byte{2}
	CurrencyPrefix  = []byte{3}
	WarehousePrefix = []byte{4}
)

const (
	ModuleName   = "rwawallet"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
	MemStoreKey  = "mem_rwawallet"
)

func WalletKey(address string) []byte {
	return append(WalletPrefix, []byte(address)...)
}

func ApprovalKey(address string) []byte {
	return append(ApprovalPrefix, []byte(address)...)
}

func CurrencyKey(denom string) []byte {
	return append(CurrencyPrefix, []byte(denom)...)
}

func WarehouseKey(name string) []byte {
	return append(WarehousePrefix, []byte(name)...)
}
