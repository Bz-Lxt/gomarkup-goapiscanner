package swagger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOpenAPI3(t *testing.T) {
	raw := []byte(`{
	  "openapi":"3.0.1",
	  "paths":{
	    "/pets/{id}":{
	      "get":{
	        "parameters":[
	          {"name":"id","in":"path","schema":{"type":"integer"}},
	          {"name":"q","in":"query","schema":{"type":"string"}}
	        ],
	        "requestBody":{
	          "content":{
	            "application/json":{
	              "schema":{"type":"object","properties":{"name":{"type":"string"}}}
	            }
	          }
	        }
	      }
	    }
	  }
	}`)
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "3.0.1" || len(got.Endpoints) != 1 {
		t.Fatalf("unexpected %#v", got)
	}
	if got.Endpoints[0].Method != "GET" {
		t.Fatal(got.Endpoints[0].Method)
	}
	if len(got.Endpoints[0].Params) < 3 {
		t.Fatalf("params=%d", len(got.Endpoints[0].Params))
	}
}

func TestParseSwagger2(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "swagger2.json"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "2.0" {
		t.Fatal(got.Version)
	}
	if len(got.Endpoints) != 2 {
		t.Fatalf("eps=%d", len(got.Endpoints))
	}
}

func TestParseRejectsEmpty(t *testing.T) {
	if _, err := Parse(nil); err == nil {
		t.Fatal("expected error")
	}
	if _, err := Parse([]byte(`{"openapi":"3.0.0"}`)); err == nil {
		t.Fatal("expected missing paths")
	}
}
