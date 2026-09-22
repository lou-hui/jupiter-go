test:
	go test -race -v ./...

generate-openapi:
	oapi-codegen -package swapv1 -generate client,types ./jupiter/openapi/swap-v1-api.yaml > ./jupiter/swapv1/client.gen.go
	oapi-codegen -package swapv2 -generate client,types ./jupiter/openapi/swap-v2-api.yaml > ./jupiter/swapv2/client.gen.go
	oapi-codegen -package pricev3 -generate client,types ./jupiter/openapi/price-v3-api.yaml > ./jupiter/pricev3/client.gen.go

lint-fix:
	golangci-lint run -E gofumpt --fix ./...

lint:
	golangci-lint run -E gofumpt ./...
