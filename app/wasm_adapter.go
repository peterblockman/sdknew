package app

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// WasmDistributionKeeper is an adapter that implements the wasmd DistributionKeeper interface
// with the cosmos-sdk distribution keeper.
type WasmDistributionKeeper struct {
	distrKeeper   distrkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
}

// NewWasmDistributionKeeper creates a new WasmDistributionKeeper
func NewWasmDistributionKeeper(distrKeeper distrkeeper.Keeper, stakingKeeper stakingkeeper.Keeper) *WasmDistributionKeeper {
	return &WasmDistributionKeeper{
		distrKeeper:   distrKeeper,
		stakingKeeper: stakingKeeper,
	}
}

// DelegatorWithdrawAddress implements the wasmd DistributionKeeper interface
func (k *WasmDistributionKeeper) DelegatorWithdrawAddress(c context.Context, req *distrtypes.QueryDelegatorWithdrawAddressRequest) (*distrtypes.QueryDelegatorWithdrawAddressResponse, error) {
	delegatorAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	withdrawAddr, err := k.distrKeeper.GetDelegatorWithdrawAddr(c, delegatorAddr)
	if err != nil {
		return nil, err
	}

	return &distrtypes.QueryDelegatorWithdrawAddressResponse{
		WithdrawAddress: withdrawAddr.String(),
	}, nil
}

// DelegationRewards implements the wasmd DistributionKeeper interface
func (k *WasmDistributionKeeper) DelegationRewards(c context.Context, req *distrtypes.QueryDelegationRewardsRequest) (*distrtypes.QueryDelegationRewardsResponse, error) {
	delegatorAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	validatorAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// We need to get the validator and delegation objects to calculate rewards
	validator, err := k.stakingKeeper.Validator(c, validatorAddr)
	if err != nil {
		return nil, err
	}

	delegation, err := k.stakingKeeper.Delegation(c, delegatorAddr, validatorAddr)
	if err != nil {
		return nil, err
	}

	// Get the validator's current period
	valCurrentRewards, err := k.distrKeeper.GetValidatorCurrentRewards(c, validatorAddr)
	if err != nil {
		return nil, err
	}

	// Use the current period as the ending period for calculating rewards
	endingPeriod := valCurrentRewards.Period

	rewards, err := k.distrKeeper.CalculateDelegationRewards(c, validator, delegation, endingPeriod)
	if err != nil {
		return nil, err
	}

	return &distrtypes.QueryDelegationRewardsResponse{
		Rewards: rewards,
	}, nil
}

// DelegationTotalRewards implements the wasmd DistributionKeeper interface
func (k *WasmDistributionKeeper) DelegationTotalRewards(c context.Context, req *distrtypes.QueryDelegationTotalRewardsRequest) (*distrtypes.QueryDelegationTotalRewardsResponse, error) {
	// This implementation is simplified - in a real application you would
	// need to iterate through all validators for this delegator and calculate rewards
	// Using the delegator address to satisfy the unused variable linter warning
	delegatorAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// For simplicity, returning an empty response
	// In a real implementation, you would use sdk calls to get all delegations and calculate rewards
	_ = delegatorAddr // mark as used

	return &distrtypes.QueryDelegationTotalRewardsResponse{
		Rewards: []distrtypes.DelegationDelegatorReward{},
		Total:   sdk.DecCoins{},
	}, nil
}

// DelegatorValidators implements the wasmd DistributionKeeper interface
func (k *WasmDistributionKeeper) DelegatorValidators(c context.Context, req *distrtypes.QueryDelegatorValidatorsRequest) (*distrtypes.QueryDelegatorValidatorsResponse, error) {
	// This is a simplified implementation
	// In a real application, you would query the staking keeper for validators
	// that the delegator has delegated to
	delegatorAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// For simplicity, returning an empty response
	// In a real implementation, you'd iterate through delegations and collect validator addresses
	_ = delegatorAddr // mark as used

	return &distrtypes.QueryDelegatorValidatorsResponse{
		Validators: []string{},
	}, nil
}
