package swagger_test

import (
	"testing"

	"github.com/alkaid/goapiscanner/internal/swagger"
)

func TestParseRequestBodyRefWithoutComponents(t *testing.T) {
	raw := []byte(`{
  "openapi": "3.0.3",
  "paths": {
    "/profiles": {
      "post": {
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/Profile"}
            }
          }
        },
        "responses": {"204": {"description": "updated"}}
      }
    }
  }
}`)

	parsed, err := swagger.Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned an error instead of degrading gracefully: %v", err)
	}
	if len(parsed.Endpoints) != 1 {
		t.Fatalf("got %d endpoints, want 1", len(parsed.Endpoints))
	}
	ep := parsed.Endpoints[0]
	if len(ep.Params) != 1 || ep.Params[0].In != swagger.InBody || ep.Params[0].Name != "body" {
		t.Fatalf("request body fallback = %#v, want one generic body parameter", ep.Params)
	}
}
