test:
	go test -race -v ./...

generate-openapi:
	oapi-codegen -package jupiter -generate client,types ./jupiter/openapi/swap-api.yaml > ./jupiter/client.gen.go
	oapi-codegen -package jupiter -generate client,types ./jupiter/openapi/swap-v2-api.yaml > ./jupiter/client_v2.gen.go

lint-fix:
	golangci-lint run -E gofumpt --fix ./...

lint:
	golangci-lint run -E gofumpt ./...
