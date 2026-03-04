generate:
	oapi-codegen -package api -generate types,chi-server -o internal/api/gen.go openapi/openapi.yaml
