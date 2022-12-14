# oapi-codegen

Taken from https://mixi-developers.mixi.co.jp/openapi-with-bazel-990fa42e8745 (Japanese)

Copy main code from https://github.com/deepmap/oapi-codegen/blob/v1.12.4/cmd/oapi-codegen/oapi-codegen.go for Bazel with Go vendoring.

```
wget -P cmd/oapi-codegen https://raw.githubusercontent.com/deepmap/oapi-codegen/v1.12.4/cmd/oapi-codegen/oapi-codegen.go
go get github.com/deepmap/oapi-codegen
go get gopkg.in/yaml.v2
go mod vendor
gazelle update -external vendored
```