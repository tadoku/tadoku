//go:generate go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4
//go:generate oapi-codegen -package openapi -generate types,server -o api.gen.go api.yaml

package openapi
