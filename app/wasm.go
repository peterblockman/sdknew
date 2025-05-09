package app

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	runtime "github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cast"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// registerWasmModule registers the wasm module and its dependencies
func (app *App) registerWasmModule(appOpts servertypes.AppOptions) error {
	// Register the store key if not already registered by IBC
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(wasmtypes.StoreKey),
	); err != nil {
		return err
	}

	// Register wasm interfaces with the InterfaceRegistry
	wasmtypes.RegisterInterfaces(app.interfaceRegistry)

	// Configure the wasm subspace
	app.ParamsKeeper.Subspace(wasmtypes.ModuleName)

	// Read wasm configuration
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		return fmt.Errorf("error reading wasm config: %w", err)
	}

	// Check for custom wasm file size limits from app options
	if maxSize := cast.ToInt(appOpts.Get("wasm.max_wasm_size")); maxSize > 0 {
		wasmtypes.MaxWasmSize = maxSize
	}

	if maxProposalSize := cast.ToInt(appOpts.Get("wasm.max_proposal_wasm_size")); maxProposalSize > 0 {
		wasmtypes.MaxProposalWasmSize = maxProposalSize
	}

	// Create a scoped keeper for wasm
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.ScopedWasmKeeper = scopedWasmKeeper

	// Use store adapter from runtime
	storeService := runtime.NewKVStoreService(app.GetKey(wasmtypes.StoreKey))

	wasmOpts := []wasmkeeper.Option{
		wasmkeeper.WithGasRegister(wasmtypes.NewDefaultWasmGasRegister()),
	}

	// Create wasm distribution keeper adapter
	wasmDistributionKeeper := NewWasmDistributionKeeper(app.DistrKeeper, *app.StakingKeeper)

	app.WasmKeeper = wasmkeeper.NewKeeper(
		app.appCodec,
		storeService,
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		wasmDistributionKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.ScopedWasmKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		"", // tempDir - leave empty for production
		wasmConfig,
		"iterator,staking,stargate", // availableCapabilities - enable commonly required capabilities
		authtypes.NewModuleAddress(govtypes.ModuleName).String(), // authority
		wasmOpts...,
	)

	// Register the wasm module with modules list
	wasmModule := wasm.NewAppModule(
		app.appCodec,
		&app.WasmKeeper,
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.MsgServiceRouter(),
		app.GetSubspace(wasmtypes.ModuleName),
	)
	if err := app.RegisterModules(wasmModule); err != nil {
		return err
	}

	return nil
}
