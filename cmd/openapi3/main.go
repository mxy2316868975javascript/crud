package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// main 读取 Swagger 2.0 文件并转换成 OpenAPI 3.0。
func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "用法: openapi3 <swagger.json> <openapi.json>")
		os.Exit(2)
	}

	swaggerContents, err := os.ReadFile(os.Args[1])
	if err != nil {
		fatal(err)
	}

	var swaggerDocument openapi2.T
	if err := json.Unmarshal(swaggerContents, &swaggerDocument); err != nil {
		fatal(fmt.Errorf("parse swagger: %w", err))
	}

	openAPIDocument, err := openapi2conv.ToV3(&swaggerDocument)
	if err != nil {
		fatal(fmt.Errorf("convert swagger: %w", err))
	}
	openAPIDocument.OpenAPI = "3.0.3"

	baseURL := os.Getenv("PUBLIC_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	openAPIDocument.Servers = openapi3.Servers{{
		URL:         baseURL,
		Description: "API server",
	}}
	patchV3(openAPIDocument)
	if err := openAPIDocument.Validate(context.Background()); err != nil {
		fatal(fmt.Errorf("validate openapi: %w", err))
	}

	output, err := json.MarshalIndent(openAPIDocument, "", "  ")
	if err != nil {
		fatal(err)
	}
	output = append(output, '\n')
	if err := os.WriteFile(os.Args[2], output, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s\n", os.Args[2])
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// 补充 Swagger 2.0 无法表达的 3.0 字段信息。
func patchV3(document *openapi3.T) {
	if document.Components == nil {
		return
	}
	for schemaName, schemaReference := range document.Components.Schemas {
		if schemaName != "Test" && !strings.HasSuffix(schemaName, ".Test") {
			continue
		}
		if schemaReference.Value == nil {
			continue
		}
		for _, fieldName := range []string{"correct_count", "estimated_vocabulary", "completed_at"} {
			if property, ok := schemaReference.Value.Properties[fieldName]; ok && property.Value != nil {
				property.Value.Nullable = true
			}
		}
	}
}
