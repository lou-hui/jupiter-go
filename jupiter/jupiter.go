// Package jupiter provides a facade over the Jupiter API, which is split
// across categories (swap, price, ...) and versions (v1, v2, v3, ...), each
// served from its own path under a common host, e.g.:
//
//	https://api.jup.ag/swap/v1/quote
//	https://api.jup.ag/swap/v2/build
//	https://api.jup.ag/price/v3
//
// Client holds the root host and knows how to build a category+version
// specific client on demand.
package jupiter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lou-hui/jupiter-go/jupiter/pricev3"
	"github.com/lou-hui/jupiter-go/jupiter/swapv1"
	"github.com/lou-hui/jupiter-go/jupiter/swapv2"
)

// DefaultAPIURL is the root host of the official Jupiter API.
// For more info visit: https://dev.jup.ag/docs
const DefaultAPIURL = "https://api.jup.ag"

// Category identifies a Jupiter API category, e.g. "swap" or "price".
type Category string

const (
	CategorySwap  Category = "swap"
	CategoryPrice Category = "price"
)

// Version identifies a Jupiter API version within a category.
type Version string

const (
	V1 Version = "v1"
	V2 Version = "v2"
	V3 Version = "v3"
)

// BuildURL joins a root host with a category and version into a full
// endpoint URL, e.g. BuildURL("https://api.jup.ag", CategorySwap, V2) ->
// "https://api.jup.ag/swap/v2".
func BuildURL(root string, category Category, version Version) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(root, "/"), category, version)
}

// HttpRequestDoer is satisfied by *http.Client and any custom HTTP client
// used by the generated sub-clients.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// RequestEditorFn mutates an outgoing HTTP request before it is sent. It
// is applied to requests made by every category/version client.
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// ClientOption configures a Client.
type ClientOption func(*Client) error

// WithHTTPClient overrides the default HTTP client used by every
// category/version client built from this Client.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.httpClient = doer
		return nil
	}
}

// WithRequestEditorFn registers a callback invoked on every outgoing
// request, for every category/version client built from this Client.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.requestEditors = append(c.requestEditors, fn)
		return nil
	}
}

// Client is the root Jupiter API client. It does not perform requests
// itself: it builds and caches category/version specific clients (SwapV1,
// SwapV2, PriceV3, ...) that share the same root host, HTTP client, and
// request editors.
type Client struct {
	root           string
	httpClient     HttpRequestDoer
	requestEditors []RequestEditorFn

	swapV1  *swapv1.ClientWithResponses
	swapV2  *swapv2.ClientWithResponses
	priceV3 *pricev3.ClientWithResponses
}

// NewClient creates a Jupiter API client rooted at the given host, e.g.
// "https://api.jup.ag" or "https://lite-api.jup.ag".
func NewClient(root string, opts ...ClientOption) (*Client, error) {
	c := &Client{root: root}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// SwapV1 returns the client for the swap/v1 category, creating it on first
// use.
func (c *Client) SwapV1() (*swapv1.ClientWithResponses, error) {
	if c.swapV1 != nil {
		return c.swapV1, nil
	}

	client, err := swapv1.NewClientWithResponses(BuildURL(c.root, CategorySwap, V1), c.swapV1Options()...)
	if err != nil {
		return nil, fmt.Errorf("build swap v1 client: %w", err)
	}

	c.swapV1 = client
	return c.swapV1, nil
}

// SwapV2 returns the client for the swap/v2 category, creating it on first
// use.
func (c *Client) SwapV2() (*swapv2.ClientWithResponses, error) {
	if c.swapV2 != nil {
		return c.swapV2, nil
	}

	client, err := swapv2.NewClientWithResponses(BuildURL(c.root, CategorySwap, V2), c.swapV2Options()...)
	if err != nil {
		return nil, fmt.Errorf("build swap v2 client: %w", err)
	}

	c.swapV2 = client
	return c.swapV2, nil
}

// PriceV3 returns the client for the price/v3 category, creating it on
// first use.
func (c *Client) PriceV3() (*pricev3.ClientWithResponses, error) {
	if c.priceV3 != nil {
		return c.priceV3, nil
	}

	client, err := pricev3.NewClientWithResponses(BuildURL(c.root, CategoryPrice, V3), c.priceV3Options()...)
	if err != nil {
		return nil, fmt.Errorf("build price v3 client: %w", err)
	}

	c.priceV3 = client
	return c.priceV3, nil
}

func (c *Client) swapV1Options() []swapv1.ClientOption {
	var opts []swapv1.ClientOption
	if c.httpClient != nil {
		opts = append(opts, swapv1.WithHTTPClient(c.httpClient))
	}
	for _, fn := range c.requestEditors {
		opts = append(opts, swapv1.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			return fn(ctx, req)
		}))
	}
	return opts
}

func (c *Client) swapV2Options() []swapv2.ClientOption {
	var opts []swapv2.ClientOption
	if c.httpClient != nil {
		opts = append(opts, swapv2.WithHTTPClient(c.httpClient))
	}
	for _, fn := range c.requestEditors {
		opts = append(opts, swapv2.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			return fn(ctx, req)
		}))
	}
	return opts
}

func (c *Client) priceV3Options() []pricev3.ClientOption {
	var opts []pricev3.ClientOption
	if c.httpClient != nil {
		opts = append(opts, pricev3.WithHTTPClient(c.httpClient))
	}
	for _, fn := range c.requestEditors {
		opts = append(opts, pricev3.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			return fn(ctx, req)
		}))
	}
	return opts
}

// The methods below forward to the category/version specific clients so
// that every Jupiter API operation is reachable from a single Client,
// without needing to call SwapV1/SwapV2/PriceV3 first.

// QuoteGetWithResponse calls GET /swap/v1/quote.
func (c *Client) QuoteGetWithResponse(ctx context.Context, params *swapv1.QuoteGetParams, reqEditors ...swapv1.RequestEditorFn) (*swapv1.QuoteGetResponse, error) {
	client, err := c.SwapV1()
	if err != nil {
		return nil, err
	}
	return client.QuoteGetWithResponse(ctx, params, reqEditors...)
}

// SwapPostWithResponse calls POST /swap/v1/swap.
func (c *Client) SwapPostWithResponse(ctx context.Context, body swapv1.SwapPostJSONRequestBody, reqEditors ...swapv1.RequestEditorFn) (*swapv1.SwapPostResponse, error) {
	client, err := c.SwapV1()
	if err != nil {
		return nil, err
	}
	return client.SwapPostWithResponse(ctx, body, reqEditors...)
}

// SwapPostWithBodyWithResponse calls POST /swap/v1/swap with a raw request body.
func (c *Client) SwapPostWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...swapv1.RequestEditorFn) (*swapv1.SwapPostResponse, error) {
	client, err := c.SwapV1()
	if err != nil {
		return nil, err
	}
	return client.SwapPostWithBodyWithResponse(ctx, contentType, body, reqEditors...)
}

// SwapInstructionsPostWithResponse calls POST /swap/v1/swap-instructions.
func (c *Client) SwapInstructionsPostWithResponse(ctx context.Context, body swapv1.SwapInstructionsPostJSONRequestBody, reqEditors ...swapv1.RequestEditorFn) (*swapv1.SwapInstructionsPostResponse, error) {
	client, err := c.SwapV1()
	if err != nil {
		return nil, err
	}
	return client.SwapInstructionsPostWithResponse(ctx, body, reqEditors...)
}

// SwapInstructionsPostWithBodyWithResponse calls POST /swap/v1/swap-instructions with a raw request body.
func (c *Client) SwapInstructionsPostWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...swapv1.RequestEditorFn) (*swapv1.SwapInstructionsPostResponse, error) {
	client, err := c.SwapV1()
	if err != nil {
		return nil, err
	}
	return client.SwapInstructionsPostWithBodyWithResponse(ctx, contentType, body, reqEditors...)
}

// ProgramIdToLabelGetWithResponse calls GET /swap/v1/program-id-to-label.
func (c *Client) ProgramIdToLabelGetWithResponse(ctx context.Context, reqEditors ...swapv1.RequestEditorFn) (*swapv1.ProgramIdToLabelGetResponse, error) {
	client, err := c.SwapV1()
	if err != nil {
		return nil, err
	}
	return client.ProgramIdToLabelGetWithResponse(ctx, reqEditors...)
}

// BuildGetWithResponse calls GET /swap/v2/build.
func (c *Client) BuildGetWithResponse(ctx context.Context, params *swapv2.BuildGetParams, reqEditors ...swapv2.RequestEditorFn) (*swapv2.BuildGetResponse, error) {
	client, err := c.SwapV2()
	if err != nil {
		return nil, err
	}
	return client.BuildGetWithResponse(ctx, params, reqEditors...)
}

// ProgramIdToLabelGetV2WithResponse calls GET /swap/v2/program-id-to-label.
func (c *Client) ProgramIdToLabelGetV2WithResponse(ctx context.Context, reqEditors ...swapv2.RequestEditorFn) (*swapv2.ProgramIdToLabelGetResponse, error) {
	client, err := c.SwapV2()
	if err != nil {
		return nil, err
	}
	return client.ProgramIdToLabelGetWithResponse(ctx, reqEditors...)
}

// GetPriceWithResponse calls GET /price/v3.
func (c *Client) GetPriceWithResponse(ctx context.Context, params *pricev3.GetPriceParams, reqEditors ...pricev3.RequestEditorFn) (*pricev3.GetPriceResponse, error) {
	client, err := c.PriceV3()
	if err != nil {
		return nil, err
	}
	return client.GetPriceWithResponse(ctx, params, reqEditors...)
}
