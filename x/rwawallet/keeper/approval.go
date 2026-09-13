package keeper

import (
	"context"
	"time"

	rwatype "github.com/babylonlabs-io/babylon/v4/x/rwawallet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) ApproveWallet(ctx context.Context, authority, walletAddress string, approvedAt time.Time) (rwatype.Approval, error) {
	if authority != k.authority {
		return rwatype.Approval{}, rwatype.ErrUnauthorized.Wrapf("expected %s", k.authority)
	}
	wallet, found, err := k.getWallet(ctx, walletAddress)
	if err != nil {
		return rwatype.Approval{}, err
	}
	if !found {
		return rwatype.Approval{}, rwatype.ErrWalletNotFound.Wrap(walletAddress)
	}
	if approvedAt.IsZero() {
		approvedAt = time.Now().UTC()
	}
	approval, found, err := k.getApproval(ctx, walletAddress)
	if err != nil {
		return rwatype.Approval{}, err
	}
	if !found {
		approval = rwatype.Approval{WalletAddress: walletAddress, RequestedBy: wallet.Owner, RequestedAt: approvedAt}
	}
	approval.Approved = true
	approval.ApprovedBy = authority
	approval.ApprovedAt = &approvedAt
	approval.RevokedAt = nil
	if err := k.setApproval(ctx, approval); err != nil {
		return rwatype.Approval{}, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		rwatype.EventTypeWalletApproved,
		sdk.NewAttribute(rwatype.AttributeKeyWallet, walletAddress),
		sdk.NewAttribute(rwatype.AttributeKeyAuthority, authority),
	))
	return approval, nil
}

func (k Keeper) RevokeApproval(ctx context.Context, authority, walletAddress string, revokedAt time.Time) (rwatype.Approval, error) {
	if authority != k.authority {
		return rwatype.Approval{}, rwatype.ErrUnauthorized.Wrapf("expected %s", k.authority)
	}
	approval, found, err := k.getApproval(ctx, walletAddress)
	if err != nil {
		return rwatype.Approval{}, err
	}
	if !found {
		return rwatype.Approval{}, rwatype.ErrApprovalRecordNotFound.Wrap(walletAddress)
	}
	if revokedAt.IsZero() {
		revokedAt = time.Now().UTC()
	}
	approval.Approved = false
	approval.RevokedAt = &revokedAt
	approval.ApprovedAt = nil
	approval.ApprovedBy = ""
	if err := k.setApproval(ctx, approval); err != nil {
		return rwatype.Approval{}, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		rwatype.EventTypeApprovalRevoked,
		sdk.NewAttribute(rwatype.AttributeKeyWallet, walletAddress),
		sdk.NewAttribute(rwatype.AttributeKeyAuthority, authority),
	))
	return approval, nil
}

func (k Keeper) GetApproval(ctx context.Context, walletAddress string) (rwatype.Approval, bool, error) {
	return k.getApproval(ctx, walletAddress)
}

func (k Keeper) QueryApproval(ctx context.Context, walletAddress string) (rwatype.Approval, bool, error) {
	return k.GetApproval(ctx, walletAddress)
}

func (k Keeper) ListApprovals(ctx context.Context) ([]rwatype.Approval, error) {
	return listByPrefix[rwatype.Approval](ctx, k.storeService, rwatype.ApprovalPrefix)
}

func (k Keeper) IsWalletApproved(ctx context.Context, walletAddress string) (bool, error) {
	approval, found, err := k.getApproval(ctx, walletAddress)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	return approval.Approved, nil
}
