package httpapi

import (
	"embed"
	"net/http"
)

//go:generate go tool swag init -g router.go -o docs --outputTypes json --parseInternal --parseDependency --parseDependencyLevel 1
//go:generate go run github.com/mxy2316868975javascript/crud/cmd/openapi3 docs/swagger.json docs/openapi.json

// 文档文件在构建时嵌入二进制，运行时不需要读磁盘。
//
//go:embed docs/swagger.json docs/openapi.json docs/docs.html
var docsFS embed.FS

func (api *API) swaggerSpec(response http.ResponseWriter, request *http.Request) {
	contents, err := docsFS.ReadFile("docs/swagger.json")
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal server error")
		return
	}
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = response.Write(contents)
}

func (api *API) openapiSpec(response http.ResponseWriter, request *http.Request) {
	contents, err := docsFS.ReadFile("docs/openapi.json")
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal server error")
		return
	}
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = response.Write(contents)
}

func (api *API) swaggerUI(response http.ResponseWriter, request *http.Request) {
	contents, err := docsFS.ReadFile("docs/docs.html")
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal server error")
		return
	}
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = response.Write(contents)
}
