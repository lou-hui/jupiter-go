package jupiter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lou-hui/jupiter-go/jupiter/pricev3"
	"github.com/lou-hui/jupiter-go/jupiter/swapv1"
	"github.com/lou-hui/jupiter-go/jupiter/swapv2"
)

func TestBuildURL(t *testing.T) {
	require.Equal(t, "https://api.jup.ag/swap/v1", BuildURL("https://api.jup.ag", CategorySwap, V1))
	require.Equal(t, "https://api.jup.ag/swap/v2", BuildURL("https://api.jup.ag", CategorySwap, V2))
	require.Equal(t, "https://api.jup.ag/price/v3", BuildURL("https://api.jup.ag", CategoryPrice, V3))
	require.Equal(t, "https://api.jup.ag/swap/v1", BuildURL("https://api.jup.ag/", CategorySwap, V1))
}

func TestClient_SubClients(t *testing.T) {
	client, err := NewClient(DefaultAPIURL)
	require.NoError(t, err)

	swapV1, err := client.SwapV1()
	require.NoError(t, err)
	require.NotNil(t, swapV1)

	swapV2, err := client.SwapV2()
	require.NoError(t, err)
	require.NotNil(t, swapV2)

	priceV3, err := client.PriceV3()
	require.NoError(t, err)
	require.NotNil(t, priceV3)

	// Sub-clients are cached: repeated calls return the same instance.
	swapV1Again, err := client.SwapV1()
	require.NoError(t, err)
	require.Same(t, swapV1, swapV1Again)
}

// TestClient_ForwardingMethods verifies that every operation is reachable
// directly from the root Client and routes to the correct category/version
// path on the underlying host.
func TestClient_ForwardingMethods(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.QuoteGetWithResponse(ctx, &swapv1.QuoteGetParams{
		InputMint:  "in",
		OutputMint: "out",
		Amount:     1,
	})
	require.NoError(t, err)
	require.Equal(t, "/swap/v1/quote", gotPath)

	_, err = client.ProgramIdToLabelGetWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, "/swap/v1/program-id-to-label", gotPath)

	_, err = client.BuildGetWithResponse(ctx, &swapv2.BuildGetParams{
		InputMint:  "in",
		OutputMint: "out",
		Amount:     "1",
		Taker:      "taker",
	})
	require.NoError(t, err)
	require.Equal(t, "/swap/v2/build", gotPath)

	_, err = client.ProgramIdToLabelGetV2WithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, "/swap/v2/program-id-to-label", gotPath)

	_, err = client.GetPriceWithResponse(ctx, &pricev3.GetPriceParams{Ids: "mint1,mint2"})
	require.NoError(t, err)
	require.Equal(t, "/price/v3/", gotPath)
}
