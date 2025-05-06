package sdknew_test

import (
	"testing"

	keepertest "sdknew/testutil/keeper"
	"sdknew/testutil/nullify"
	sdknew "sdknew/x/sdknew/module"
	"sdknew/x/sdknew/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.SdknewKeeper(t)
	sdknew.InitGenesis(ctx, k, genesisState)
	got := sdknew.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
