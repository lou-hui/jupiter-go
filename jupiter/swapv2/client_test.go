package swapv2

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/buildResponse.json
var buildResponseJSON []byte

func TestBuildResponse_Unmarshal(t *testing.T) {
	var response BuildResponse

	err := json.Unmarshal(buildResponseJSON, &response)
	require.NoError(t, err)

	require.NotNil(t, response.InputMint)
	require.Equal(t, "So11111111111111111111111111111111111111112", *response.InputMint)

	require.NotNil(t, response.OutputMint)
	require.Equal(t, "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", *response.OutputMint)

	require.NotNil(t, response.InAmount)
	require.Equal(t, "1000000", *response.InAmount)

	require.NotNil(t, response.RoutePlan)
	require.Len(t, *response.RoutePlan, 1)

	require.NotNil(t, response.ComputeBudgetInstructions)
	require.Len(t, *response.ComputeBudgetInstructions, 1)

	require.NotNil(t, response.SetupInstructions)
	require.Len(t, *response.SetupInstructions, 1)

	require.NotNil(t, response.SwapInstruction)

	require.NotNil(t, response.CleanupInstruction)

	require.NotNil(t, response.OtherInstructions)
	require.Empty(t, *response.OtherInstructions)

	require.Nil(t, response.TipInstruction)

	require.NotNil(t, response.AddressesByLookupTableAddress)
	require.Len(t, *response.AddressesByLookupTableAddress, 1)

	require.NotNil(t, response.BlockhashWithMetadata)
	require.Equal(t, 299999999, *response.BlockhashWithMetadata.LastValidBlockHeight)
	require.NotNil(t, response.BlockhashWithMetadata.Blockhash)
	require.Len(t, *response.BlockhashWithMetadata.Blockhash, 32)
	require.NotNil(t, response.BlockhashWithMetadata.FetchedAt)
}
