package app_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"sdknew/app"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func TestWasmModuleWithTraceStore(t *testing.T) {
	// Setup a logger and in-memory database
	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()

	// Create a bytes buffer to capture trace output
	traceOutput := &bytes.Buffer{}

	// Create app options
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = 1

	// Initialize the application with the trace store
	testApp, err := app.New(logger, db, traceOutput, true, appOptions)
	require.NoError(t, err, "Failed to create application with trace store")

	// Verify that WasmKeeper is initialized
	require.NotNil(t, testApp.WasmKeeper, "WasmKeeper should be initialized with trace store")

	// Check that the wasm store key exists
	storeKey := testApp.GetKey(wasmtypes.StoreKey)
	require.NotNil(t, storeKey, "Wasm store key should exist with trace store")

	/*
		The trace operations require block height and transaction hash metadata to generate proper output.
	*/
	// Create a context manually with proper setup
	stateStore := testApp.CommitMultiStore()
	header := cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	ctx := sdk.NewContext(stateStore, header, false, testApp.Logger())
	ctx = ctx.WithBlockHeight(1).
		WithTxBytes([]byte("test_tx_bytes")).
		WithVoteInfos([]abci.VoteInfo{})

	// Access the store through the context to ensure all operations are traced
	store := ctx.KVStore(storeKey)

	// Perform multiple operations to generate trace data
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("test-key-%d", i))
		value := []byte(fmt.Sprintf("test-value-%d", i))
		store.Set(key, value)
	}

	// Delete one of the keys to generate another type of operation
	store.Delete([]byte("test-key-5"))

	// Retrieve some values
	for i := 0; i < 10; i++ {
		if i == 5 {
			continue // We deleted this one
		}
		key := []byte(fmt.Sprintf("test-key-%d", i))
		_ = store.Get(key)
	}

	// We don't need to commit changes for trace data to be captured

	// Check if something was captured in the trace output
	t.Logf("Trace data captured: %d bytes", traceOutput.Len())
	require.NotZero(t, traceOutput.Len(), "Trace output should contain data")

	// Log a small sample of the trace data for debugging purposes
	if traceOutput.Len() > 0 {
		sampleSize := 200
		if traceOutput.Len() < sampleSize {
			sampleSize = traceOutput.Len()
		}
		t.Logf("Sample trace data: %s", traceOutput.String()[:sampleSize])
	}
}
