package types

import (
	"fmt"
	"sort"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CurrencyStateActive   = "active"
	CurrencyStateInactive = "inactive"
	CurrencyStateRotated  = "rotated"
)

type Wallet struct {
	Address            string              `json:"address"`
	Owner              string              `json:"owner"`
	Warehouse          string              `json:"warehouse,omitempty"`
	Balances           sdk.Coins           `json:"balances"`
	TransactionHistory []TransactionRecord `json:"transaction_history"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

type TransactionRecord struct {
	ID         string    `json:"id"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Amount     sdk.Coin  `json:"amount"`
	Cycle      uint64    `json:"cycle"`
	State      string    `json:"state"`
	ApprovedBy string    `json:"approved_by"`
	Executor   string    `json:"executor"`
	Timestamp  time.Time `json:"timestamp"`
}

type Approval struct {
	WalletAddress string     `json:"wallet_address"`
	RequestedBy   string     `json:"requested_by"`
	Approved      bool       `json:"approved"`
	ApprovedBy    string     `json:"approved_by,omitempty"`
	RequestedAt   time.Time  `json:"requested_at"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
}

type Currency struct {
	Denom               string        `json:"denom"`
	State               string        `json:"state"`
	Cycle               uint64        `json:"cycle"`
	CycleDuration       time.Duration `json:"cycle_duration"`
	LastRotatedAt       time.Time     `json:"last_rotated_at"`
	SupportedWarehouses []string      `json:"supported_warehouses,omitempty"`
}

type Warehouse struct {
	Name            string    `json:"name"`
	WalletAddress   string    `json:"wallet_address"`
	SupportedDenoms []string  `json:"supported_denoms,omitempty"`
	InitializedAt   time.Time `json:"initialized_at"`
}

type GenesisState struct {
	Authority  string      `json:"authority"`
	Wallets    []Wallet    `json:"wallets"`
	Approvals  []Approval  `json:"approvals"`
	Currencies []Currency  `json:"currencies"`
	Warehouses []Warehouse `json:"warehouses"`
}

func DefaultGenesis(authority string) *GenesisState {
	return &GenesisState{
		Authority:  authority,
		Wallets:    []Wallet{},
		Approvals:  []Approval{},
		Currencies: []Currency{},
		Warehouses: []Warehouse{},
	}
}

func (gs GenesisState) Validate() error {
	seenWallets := map[string]struct{}{}
	walletWarehouses := map[string]string{}
	for _, wallet := range gs.Wallets {
		if err := wallet.Validate(); err != nil {
			return err
		}
		if _, ok := seenWallets[wallet.Address]; ok {
			return ErrDuplicateGenesisEntry.Wrapf("wallet %s", wallet.Address)
		}
		seenWallets[wallet.Address] = struct{}{}
		walletWarehouses[wallet.Address] = wallet.Warehouse
	}

	seenApprovals := map[string]struct{}{}
	for _, approval := range gs.Approvals {
		if err := approval.Validate(); err != nil {
			return err
		}
		if _, ok := seenApprovals[approval.WalletAddress]; ok {
			return ErrDuplicateGenesisEntry.Wrapf("approval %s", approval.WalletAddress)
		}
		if _, ok := seenWallets[approval.WalletAddress]; !ok {
			return ErrWalletNotFound.Wrapf("approval wallet %s", approval.WalletAddress)
		}
		seenApprovals[approval.WalletAddress] = struct{}{}
	}

	seenCurrencies := map[string]struct{}{}
	for _, currency := range gs.Currencies {
		if err := currency.Validate(); err != nil {
			return err
		}
		if _, ok := seenCurrencies[currency.Denom]; ok {
			return ErrDuplicateGenesisEntry.Wrapf("currency %s", currency.Denom)
		}
		seenCurrencies[currency.Denom] = struct{}{}
	}

	seenWarehouses := map[string]struct{}{}
	for _, warehouse := range gs.Warehouses {
		if err := warehouse.Validate(); err != nil {
			return err
		}
		if _, ok := seenWarehouses[warehouse.Name]; ok {
			return ErrDuplicateGenesisEntry.Wrapf("warehouse %s", warehouse.Name)
		}
		if _, ok := seenWallets[warehouse.WalletAddress]; !ok {
			return ErrWalletNotFound.Wrapf("warehouse wallet %s", warehouse.WalletAddress)
		}
		seenWarehouses[warehouse.Name] = struct{}{}
	}

	for walletAddress, warehouseName := range walletWarehouses {
		if warehouseName == "" {
			continue
		}
		if _, ok := seenWarehouses[warehouseName]; !ok {
			return ErrWarehouseNotFound.Wrapf("wallet %s warehouse %s", walletAddress, warehouseName)
		}
	}

	for _, currency := range gs.Currencies {
		for _, warehouseName := range currency.SupportedWarehouses {
			if _, ok := seenWarehouses[warehouseName]; !ok {
				return ErrWarehouseNotFound.Wrapf("currency %s warehouse %s", currency.Denom, warehouseName)
			}
		}
	}

	return nil
}

func (w Wallet) Validate() error {
	if strings.TrimSpace(w.Address) == "" {
		return fmt.Errorf("wallet address cannot be empty")
	}
	if strings.TrimSpace(w.Owner) == "" {
		return fmt.Errorf("wallet owner cannot be empty")
	}
	if !w.Balances.IsValid() {
		return fmt.Errorf("invalid wallet balances")
	}
	for _, tx := range w.TransactionHistory {
		if err := tx.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a Approval) Validate() error {
	if strings.TrimSpace(a.WalletAddress) == "" {
		return fmt.Errorf("approval wallet address cannot be empty")
	}
	if strings.TrimSpace(a.RequestedBy) == "" {
		return fmt.Errorf("approval requester cannot be empty")
	}
	return nil
}

func (c Currency) Validate() error {
	if err := sdk.ValidateDenom(strings.TrimSpace(c.Denom)); err != nil {
		return err
	}
	if !IsValidCurrencyState(c.State) {
		return ErrInvalidCurrencyState.Wrap(c.State)
	}
	return nil
}

func (w Warehouse) Validate() error {
	if strings.TrimSpace(w.Name) == "" {
		return fmt.Errorf("warehouse name cannot be empty")
	}
	if strings.TrimSpace(w.WalletAddress) == "" {
		return fmt.Errorf("warehouse wallet address cannot be empty")
	}
	for _, denom := range w.SupportedDenoms {
		if err := sdk.ValidateDenom(strings.TrimSpace(denom)); err != nil {
			return err
		}
	}
	return nil
}

func (t TransactionRecord) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transaction id cannot be empty")
	}
	if strings.TrimSpace(t.From) == "" || strings.TrimSpace(t.To) == "" {
		return fmt.Errorf("transaction endpoints cannot be empty")
	}
	if !t.Amount.IsValid() || !t.Amount.Amount.IsPositive() {
		return fmt.Errorf("transaction amount must be positive")
	}
	if !IsValidCurrencyState(t.State) {
		return ErrInvalidCurrencyState.Wrap(t.State)
	}
	return nil
}

func IsValidCurrencyState(state string) bool {
	switch state {
	case CurrencyStateActive, CurrencyStateInactive, CurrencyStateRotated:
		return true
	default:
		return false
	}
}

func NextCurrencyState(state string) (string, error) {
	switch state {
	case CurrencyStateActive:
		return CurrencyStateInactive, nil
	case CurrencyStateInactive:
		return CurrencyStateRotated, nil
	case CurrencyStateRotated:
		return CurrencyStateActive, nil
	default:
		return "", ErrInvalidCurrencyState.Wrap(state)
	}
}

func SortData(gs *GenesisState) {
	sort.Slice(gs.Wallets, func(i, j int) bool {
		return gs.Wallets[i].Address < gs.Wallets[j].Address
	})
	sort.Slice(gs.Approvals, func(i, j int) bool {
		return gs.Approvals[i].WalletAddress < gs.Approvals[j].WalletAddress
	})
	sort.Slice(gs.Currencies, func(i, j int) bool {
		return gs.Currencies[i].Denom < gs.Currencies[j].Denom
	})
	sort.Slice(gs.Warehouses, func(i, j int) bool {
		return gs.Warehouses[i].Name < gs.Warehouses[j].Name
	})
}
