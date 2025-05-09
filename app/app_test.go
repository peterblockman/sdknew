package app_test

import (
	"testing"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"sdknew/app"
)

func TestWasmModuleInitialization(t *testing.T) {
	// Setup a logger and in-memory database
	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()

	// Create app options
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = 1

	// Initialize the application
	testApp, err := app.New(logger, db, nil, true, appOptions)
	require.NoError(t, err, "Failed to create application")

	// Verify that WasmKeeper is initialized
	require.NotNil(t, testApp.WasmKeeper, "WasmKeeper should be initialized")

	// Check that the wasm store key exists
	storeKey := testApp.GetKey(wasmtypes.StoreKey)
	require.NotNil(t, storeKey, "Wasm store key should exist")

	// Additional checks specific to wasm module configuration
	paramSpace := testApp.GetSubspace(wasmtypes.ModuleName)
	require.NotNil(t, paramSpace, "Wasm param space should exist")

	// Check if Wasm module is registered in the module manager
	modules := testApp.ModuleManager.Modules
	_, exists := modules[wasmtypes.ModuleName]
	require.True(t, exists, "Wasm module should be registered in ModuleManager")

	// Check the ScopedWasmKeeper
	require.NotNil(t, testApp.ScopedWasmKeeper, "ScopedWasmKeeper should be initialized")
}

func TestWasmStoreIsRegistered(t *testing.T) {
	// Setup a logger and in-memory database
	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()

	// Create app options
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = 1

	// Initialize the application
	testApp, err := app.New(logger, db, nil, true, appOptions)
	require.NoError(t, err, "Failed to create application")

	// Check that wasm store is in the multistore
	storeKeys := testApp.GetStoreKeys()
	var foundWasmKey bool
	for _, key := range storeKeys {
		if key.Name() == wasmtypes.StoreKey {
			foundWasmKey = true
			break
		}
	}
	require.True(t, foundWasmKey, "Wasm store key should be in the multistore")
}
