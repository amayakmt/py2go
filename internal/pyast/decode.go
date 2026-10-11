package pyast

import (
	"encoding/json"
	"fmt"
)

// Decode parses the JSON from dump_ast.py into a Node tree.
func Decode(data []byte) (Node, error) {
	return decodeNode(json.RawMessage(data))
}

func decodeNode(raw json.RawMessage) (Node, error) {
	// Pass 1: read only the type tag and line number.
	var head struct {
		Type   string `json:"_type"`
		Lineno int    `json:"lineno"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, fmt.Errorf("decoding node %.60s: %w", string(raw), err)
	}

	switch head.Type {
	case "Module":
		var aux struct {
			Body []json.RawMessage `json:"body"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		body, err := decodeList(aux.Body)
		if err != nil {
			return nil, err
		}
		return &Module{Body: body}, nil

	case "FunctionDef":
		var aux struct {
			Name    string            `json:"name"`
			Body    []json.RawMessage `json:"body"`
			Returns json.RawMessage   `json:"returns"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		body, err := decodeList(aux.Body)
		if err != nil {
			return nil, err
		}
		returns, err := decodeOptional(aux.Returns)
		if err != nil {
			return nil, err
		}
		return &FunctionDef{Name: aux.Name, Body: body, Returns: returns, Lineno: head.Lineno}, nil

	case "Expr":
		var aux struct {
			Value json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		value, err := decodeNode(aux.Value)
		if err != nil {
			return nil, err
		}
		return &Expr{Value: value, Lineno: head.Lineno}, nil

	case "Call":
		var aux struct {
			Func json.RawMessage   `json:"func"`
			Args []json.RawMessage `json:"args"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		fn, err := decodeNode(aux.Func)
		if err != nil {
			return nil, err
		}
		args, err := decodeList(aux.Args)
		if err != nil {
			return nil, err
		}
		return &Call{Func: fn, Args: args, Lineno: head.Lineno}, nil

	case "Name":
		var aux struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		return &Name{ID: aux.ID, Lineno: head.Lineno}, nil

	case "Constant":
		var aux struct {
			Value any `json:"value"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		return &Constant{Value: aux.Value, Lineno: head.Lineno}, nil

	default:
		return &Unsupported{Type: head.Type, Lineno: head.Lineno}, nil
	}
}

func decodeList(raws []json.RawMessage) ([]Node, error) {
	nodes := make([]Node, 0, len(raws))
	for _, r := range raws {
		n, err := decodeNode(r)
		if err != nil {
			return nil, fmt.Errorf("decoding list item: %w", err)
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// decodeOptional returns nil for JSON null or a missing field.
func decodeOptional(raw json.RawMessage) (Node, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	return decodeNode(raw)
}
