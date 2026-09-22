package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lou-hui/jupiter-go/jupiter"
	"github.com/lou-hui/jupiter-go/jupiter/pricev3"
)

func main() {
	// Initialize the root client with API key (automatically added to all requests)
	apiKey := "{YOUR_JUPITER_API_KEY}"
	jupClient, err := jupiter.NewClient(
		jupiter.DefaultAPIURL,
		jupiter.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("x-api-key", apiKey)
			return nil
		}),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	ids := "So11111111111111111111111111111111111111112,EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	priceResp, err := jupClient.GetPriceWithResponse(ctx, &pricev3.GetPriceParams{
		Ids: ids,
	})
	if err != nil {
		panic(err)
	}

	if priceResp.JSON200 == nil {
		panic("invalid GetPriceWithResponse response")
	}

	prices := *priceResp.JSON200
	if raw, err := json.MarshalIndent(prices, "", "  "); err == nil {
		fmt.Println(string(raw))
	}

	for mint, price := range prices {
		fmt.Printf("%s: $%v (liquidity: %v)\n", mint, price.UsdPrice, price.Liquidity)
	}
}
