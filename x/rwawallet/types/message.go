package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Message interface {
	Validate() error
}

type MsgCreateWallet struct {
	Address         string    `json:"address"`
	Owner           string    `json:"owner"`
	Warehouse       string    `json:"warehouse,omitempty"`
	InitialBalances sdk.Coins `json:"initial_balances"`
	CreatedAt       time.Time `json:"created_at"`
}

func (m MsgCreateWallet) Validate() error {
	if strings.TrimSpace(m.Address) == "" {
		return fmt.Errorf("wallet address cannot be empty")
	}
	if strings.TrimSpace(m.Owner) == "" {
		return fmt.Errorf("wallet owner cannot be empty")
	}
	if !m.InitialBalances.IsValid() {
		return fmt.Errorf("invalid initial balances")
	}
	return nil
}

type MsgApproveWallet struct {
	Authority     string    `json:"authority"`
	WalletAddress string    `json:"wallet_address"`
	ApprovedAt    time.Time `json:"approved_at"`
}

func (m MsgApproveWallet) Validate() error {
	if strings.TrimSpace(m.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if strings.TrimSpace(m.WalletAddress) == "" {
		return fmt.Errorf("wallet address cannot be empty")
	}
	return nil
}

type MsgRevokeApproval struct {
	Authority     string    `json:"authority"`
	WalletAddress string    `json:"wallet_address"`
	RevokedAt     time.Time `json:"revoked_at"`
}

func (m MsgRevokeApproval) Validate() error {
	if strings.TrimSpace(m.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if strings.TrimSpace(m.WalletAddress) == "" {
		return fmt.Errorf("wallet address cannot be empty")
	}
	return nil
}

type MsgTransfer struct {
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      sdk.Coin  `json:"amount"`
	Executor    string    `json:"executor"`
	ExecutedAt  time.Time `json:"executed_at"`
}

func (m MsgTransfer) Validate() error {
	if strings.TrimSpace(m.FromAddress) == "" {
		return fmt.Errorf("source wallet address cannot be empty")
	}
	if strings.TrimSpace(m.ToAddress) == "" {
		return fmt.Errorf("destination wallet address cannot be empty")
	}
	if !m.Amount.IsValid() || !m.Amount.Amount.IsPositive() {
		return fmt.Errorf("transfer amount must be a valid positive coin")
	}
	if strings.TrimSpace(m.Executor) == "" {
		return fmt.Errorf("executor cannot be empty")
	}
	return nil
}

type MsgRotateCurrency struct {
	Authority string    `json:"authority"`
	Denom     string    `json:"denom"`
	RotatedAt time.Time `json:"rotated_at"`
}

func (m MsgRotateCurrency) Validate() error {
	if strings.TrimSpace(m.Authority) == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if err := sdk.ValidateDenom(strings.TrimSpace(m.Denom)); err != nil {
		return err
	}
	return nil
}
