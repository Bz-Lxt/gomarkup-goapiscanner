package swagger

import (
	"encoding/json"
	"strings"
)

type pathItem struct {
	Parameters []json.RawMessage
	Ops        map[string]json.RawMessage
}

func decodePathItem(raw json.RawMessage) (pathItem, error) {
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(raw, &generic); err != nil {
		return pathItem{}, err
	}
	item := pathItem{Ops: map[string]json.RawMessage{}}
	for k, v := range generic {
		lk := strings.ToLower(k)
		if lk == "parameters" {
			var arr []json.RawMessage
			if err := json.Unmarshal(v, &arr); err != nil {
				return pathItem{}, err
			}
			item.Parameters = arr
			continue
		}
		for _, m := range methods {
			if lk == m {
				item.Ops[m] = v
			}
		}
	}
	return item, nil
}

func decodeOperation(method, path string, raw json.RawMessage, shared []json.RawMessage, doc Document) (Endpoint, error) {
	var op map[string]json.RawMessage
	if err := json.Unmarshal(raw, &op); err != nil {
		return Endpoint{}, err
	}
	ep := Endpoint{Method: strings.ToUpper(method), Path: path}
	if v, ok := op["operationId"]; ok {
		_ = json.Unmarshal(v, &ep.OperationID)
	}
	var params []json.RawMessage
	if v, ok := op["parameters"]; ok {
		if err := json.Unmarshal(v, &params); err != nil {
			return Endpoint{}, err
		}
	}
	params = append(append([]json.RawMessage{}, shared...), params...)
	seen := map[string]struct{}{}
	for _, p := range params {
		pr, err := decodeParam(p)
		if err != nil || pr.Name == "" {
			continue
		}
		key := string(pr.In) + ":" + pr.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		ep.Params = append(ep.Params, pr)
	}
	if v, ok := op["requestBody"]; ok {
		for _, bp := range decodeRequestBody(v, doc) {
			key := string(bp.In) + ":" + bp.Name
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			ep.Params = append(ep.Params, bp)
		}
	}
	if len(ep.Params) == 0 {
		ep.Params = append(ep.Params, Param{Name: "probe", In: InQuery, Type: "string"})
	}
	return ep, nil
}

func decodeParam(raw json.RawMessage) (Param, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return Param{}, err
	}
	name, _ := m["name"].(string)
	in, _ := m["in"].(string)
	typ, _ := m["type"].(string)
	if typ == "" {
		if schema, ok := m["schema"].(map[string]any); ok {
			typ, _ = schema["type"].(string)
		}
	}
	req, _ := m["required"].(bool)
	pin := ParamIn(strings.ToLower(in))
	if pin == "" {
		pin = InQuery
	}
	if pin == "formdata" {
		pin = InBody
	}
	if typ == "" {
		typ = "string"
	}
	return Param{Name: name, In: pin, Type: typ, Required: req}, nil
}

func decodeRequestBody(raw json.RawMessage, doc Document) []Param {
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil
	}
	content, _ := body["content"].(map[string]any)
	if content == nil {
		return []Param{{Name: "body", In: InBody, Type: "object"}}
	}
	for _, media := range content {
		mm, ok := media.(map[string]any)
		if !ok {
			continue
		}
		schema, _ := mm["schema"].(map[string]any)
		props := flattenSchema(schema, doc, 0)
		if len(props) > 0 {
			return props
		}
	}
	return []Param{{Name: "body", In: InBody, Type: "object"}}
}

func flattenSchema(schema map[string]any, doc Document, depth int) []Param {
	if schema == nil || depth > 4 {
		return nil
	}
	if ref, _ := schema["$ref"].(string); ref != "" {
		return flattenSchema(resolveRef(ref, doc), doc, depth+1)
	}
	props, _ := schema["properties"].(map[string]any)
	var out []Param
	for name, raw := range props {
		pm, _ := raw.(map[string]any)
		typ, _ := pm["type"].(string)
		if typ == "" {
			typ = "string"
		}
		out = append(out, Param{Name: name, In: InBody, Type: typ})
	}
	return out
}

func resolveRef(ref string, doc Document) map[string]any {
	const prefix3 = "#/components/schemas/"
	const prefix2 = "#/definitions/"
	var raw json.RawMessage
	switch {
	case strings.HasPrefix(ref, prefix3) && doc.Components != nil:
		raw = doc.Components.Schemas[strings.TrimPrefix(ref, prefix3)]
	case strings.HasPrefix(ref, prefix2):
		raw = doc.Definitions[strings.TrimPrefix(ref, prefix2)]
	default:
		return nil
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return m
}
