package rwawallet

import (
	"context"
	"fmt"

	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/keeper"
	"github.com/babylonlabs-io/babylon/v4/x/rwawallet/types"
)

func HandleMessage(ctx context.Context, k keeper.Keeper, msg types.Message) error {
	switch m := msg.(type) {
	case types.MsgCreateWallet:
		_, err := k.CreateWallet(ctx, m)
		return err
	case types.MsgApproveWallet:
		_, err := k.ApproveWallet(ctx, m.Authority, m.WalletAddress, m.ApprovedAt)
		return err
	case types.MsgRevokeApproval:
		_, err := k.RevokeApproval(ctx, m.Authority, m.WalletAddress, m.RevokedAt)
		return err
	case types.MsgTransfer:
		_, err := k.Transfer(ctx, m)
		return err
	case types.MsgRotateCurrency:
		_, err := k.RotateCurrency(ctx, m.Authority, m.Denom, m.RotatedAt)
		return err
	default:
		return fmt.Errorf("unsupported rwa wallet message %T", msg)
	}
}
