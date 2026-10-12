package pyast

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
			Name          string            `json:"name"`
			Args          json.RawMessage   `json:"args"`
			Body          []json.RawMessage `json:"body"`
			DecoratorList []json.RawMessage `json:"decorator_list"`
			Returns       json.RawMessage   `json:"returns"`
			TypeParams    []json.RawMessage `json:"type_params"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		args, err := decodeArguments(aux.Args)
		if err != nil {
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
		hasParams := len(args.PosOnly)+len(args.Args)+len(args.Kwonly) > 0 ||
			args.Vararg != nil || args.Kwarg != nil
		return &FunctionDef{
			Name: aux.Name, Args: args, Body: body, Returns: returns,
			HasParams:     hasParams,
			HasDecorators: len(aux.DecoratorList) > 0,
			HasTypeParams: len(aux.TypeParams) > 0,
			Lineno:        head.Lineno,
		}, nil

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
			Func     json.RawMessage   `json:"func"`
			Args     []json.RawMessage `json:"args"`
			Keywords []json.RawMessage `json:"keywords"`
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
		var keywords []*Keyword
		for _, r := range aux.Keywords {
			var kw struct {
				Arg    string          `json:"arg"`
				Value  json.RawMessage `json:"value"`
				Lineno int             `json:"lineno"`
			}
			if err := json.Unmarshal(r, &kw); err != nil {
				return nil, err
			}
			value, err := decodeNode(kw.Value)
			if err != nil {
				return nil, err
			}
			keywords = append(keywords, &Keyword{Arg: kw.Arg, Value: value, Lineno: kw.Lineno})
		}
		return &Call{
			Func: fn, Args: args, Keywords: keywords,
			HasKeywords: len(keywords) > 0, Lineno: head.Lineno,
		}, nil

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
			PyType string          `json:"pytype"`
			Value  json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		return decodeConstant(aux.PyType, aux.Value, head.Lineno)

	case "Return":
		var aux struct {
			Value json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		value, err := decodeOptional(aux.Value)
		if err != nil {
			return nil, err
		}
		return &Return{Value: value, Lineno: head.Lineno}, nil

	case "Assign":
		var aux struct {
			Targets []json.RawMessage `json:"targets"`
			Value   json.RawMessage   `json:"value"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		targets, err := decodeList(aux.Targets)
		if err != nil {
			return nil, err
		}
		value, err := decodeNode(aux.Value)
		if err != nil {
			return nil, err
		}
		return &Assign{Targets: targets, Value: value, Lineno: head.Lineno}, nil

	case "AnnAssign":
		var aux struct {
			Target     json.RawMessage `json:"target"`
			Annotation json.RawMessage `json:"annotation"`
			Value      json.RawMessage `json:"value"`
			Simple     int             `json:"simple"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		target, err := decodeNode(aux.Target)
		if err != nil {
			return nil, err
		}
		ann, err := decodeNode(aux.Annotation)
		if err != nil {
			return nil, err
		}
		value, err := decodeOptional(aux.Value)
		if err != nil {
			return nil, err
		}
		return &AnnAssign{
			Target: target, Annotation: ann, Value: value,
			Simple: aux.Simple == 1, Lineno: head.Lineno,
		}, nil

	case "AugAssign":
		var aux struct {
			Target json.RawMessage `json:"target"`
			Op     json.RawMessage `json:"op"`
			Value  json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		target, err := decodeNode(aux.Target)
		if err != nil {
			return nil, err
		}
		op, err := decodeOp(aux.Op)
		if err != nil {
			return nil, err
		}
		value, err := decodeNode(aux.Value)
		if err != nil {
			return nil, err
		}
		return &AugAssign{Target: target, Op: op, Value: value, Lineno: head.Lineno}, nil

	case "If", "While":
		var aux struct {
			Test   json.RawMessage   `json:"test"`
			Body   []json.RawMessage `json:"body"`
			Orelse []json.RawMessage `json:"orelse"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		test, err := decodeNode(aux.Test)
		if err != nil {
			return nil, err
		}
		body, err := decodeList(aux.Body)
		if err != nil {
			return nil, err
		}
		orelse, err := decodeList(aux.Orelse)
		if err != nil {
			return nil, err
		}
		if head.Type == "If" {
			return &If{Test: test, Body: body, Orelse: orelse, Lineno: head.Lineno}, nil
		}
		return &While{Test: test, Body: body, Orelse: orelse, Lineno: head.Lineno}, nil

	case "For":
		var aux struct {
			Target json.RawMessage   `json:"target"`
			Iter   json.RawMessage   `json:"iter"`
			Body   []json.RawMessage `json:"body"`
			Orelse []json.RawMessage `json:"orelse"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		target, err := decodeNode(aux.Target)
		if err != nil {
			return nil, err
		}
		iter, err := decodeNode(aux.Iter)
		if err != nil {
			return nil, err
		}
		body, err := decodeList(aux.Body)
		if err != nil {
			return nil, err
		}
		orelse, err := decodeList(aux.Orelse)
		if err != nil {
			return nil, err
		}
		return &For{Target: target, Iter: iter, Body: body, Orelse: orelse, Lineno: head.Lineno}, nil

	case "Break":
		return &Break{Lineno: head.Lineno}, nil
	case "Continue":
		return &Continue{Lineno: head.Lineno}, nil
	case "Pass":
		return &Pass{Lineno: head.Lineno}, nil

	case "BinOp":
		var aux struct {
			Left  json.RawMessage `json:"left"`
			Op    json.RawMessage `json:"op"`
			Right json.RawMessage `json:"right"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		left, err := decodeNode(aux.Left)
		if err != nil {
			return nil, err
		}
		op, err := decodeOp(aux.Op)
		if err != nil {
			return nil, err
		}
		right, err := decodeNode(aux.Right)
		if err != nil {
			return nil, err
		}
		return &BinOp{Left: left, Op: op, Right: right, Lineno: head.Lineno}, nil

	case "UnaryOp":
		var aux struct {
			Op      json.RawMessage `json:"op"`
			Operand json.RawMessage `json:"operand"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		op, err := decodeOp(aux.Op)
		if err != nil {
			return nil, err
		}
		operand, err := decodeNode(aux.Operand)
		if err != nil {
			return nil, err
		}
		return &UnaryOp{Op: op, Operand: operand, Lineno: head.Lineno}, nil

	case "Compare":
		var aux struct {
			Left        json.RawMessage   `json:"left"`
			Ops         []json.RawMessage `json:"ops"`
			Comparators []json.RawMessage `json:"comparators"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		left, err := decodeNode(aux.Left)
		if err != nil {
			return nil, err
		}
		ops, err := decodeOps(aux.Ops)
		if err != nil {
			return nil, err
		}
		comps, err := decodeList(aux.Comparators)
		if err != nil {
			return nil, err
		}
		return &Compare{Left: left, Ops: ops, Comparators: comps, Lineno: head.Lineno}, nil

	case "BoolOp":
		var aux struct {
			Op     json.RawMessage   `json:"op"`
			Values []json.RawMessage `json:"values"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		op, err := decodeOp(aux.Op)
		if err != nil {
			return nil, err
		}
		values, err := decodeList(aux.Values)
		if err != nil {
			return nil, err
		}
		return &BoolOp{Op: op, Values: values, Lineno: head.Lineno}, nil

	case "Attribute":
		var aux struct {
			Value json.RawMessage `json:"value"`
			Attr  string          `json:"attr"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		value, err := decodeNode(aux.Value)
		if err != nil {
			return nil, err
		}
		return &Attribute{Value: value, Attr: aux.Attr, Lineno: head.Lineno}, nil

	case "Subscript":
		var aux struct {
			Value json.RawMessage `json:"value"`
			Slice json.RawMessage `json:"slice"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		value, err := decodeNode(aux.Value)
		if err != nil {
			return nil, err
		}
		index, err := decodeNode(aux.Slice)
		if err != nil {
			return nil, err
		}
		return &Subscript{Value: value, Index: index, Lineno: head.Lineno}, nil

	case "Slice":
		var aux struct {
			Lower json.RawMessage `json:"lower"`
			Upper json.RawMessage `json:"upper"`
			Step  json.RawMessage `json:"step"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		s := &Slice{Lineno: head.Lineno}
		var err error
		if s.Lower, err = decodeOptional(aux.Lower); err != nil {
			return nil, err
		}
		if s.Upper, err = decodeOptional(aux.Upper); err != nil {
			return nil, err
		}
		if s.Step, err = decodeOptional(aux.Step); err != nil {
			return nil, err
		}
		return s, nil

	case "List", "Tuple":
		var aux struct {
			Elts []json.RawMessage `json:"elts"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		elts, err := decodeList(aux.Elts)
		if err != nil {
			return nil, err
		}
		if head.Type == "List" {
			return &List{Elts: elts, Lineno: head.Lineno}, nil
		}
		return &Tuple{Elts: elts, Lineno: head.Lineno}, nil

	case "JoinedStr":
		var aux struct {
			Values []json.RawMessage `json:"values"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		values, err := decodeList(aux.Values)
		if err != nil {
			return nil, err
		}
		return &JoinedStr{Values: values, Lineno: head.Lineno}, nil

	case "FormattedValue":
		var aux struct {
			Value      json.RawMessage `json:"value"`
			Conversion int             `json:"conversion"`
			FormatSpec json.RawMessage `json:"format_spec"`
		}
		if err := json.Unmarshal(raw, &aux); err != nil {
			return nil, err
		}
		value, err := decodeNode(aux.Value)
		if err != nil {
			return nil, err
		}
		spec, err := decodeOptional(aux.FormatSpec)
		if err != nil {
			return nil, err
		}
		return &FormattedValue{Value: value, Conversion: aux.Conversion, FormatSpec: spec, Lineno: head.Lineno}, nil

	default:
		return &Unsupported{Type: head.Type, Lineno: head.Lineno}, nil
	}
}

// decodeOp reads an operator node like {"_type": "Add"} into "Add".
func decodeOp(raw json.RawMessage) (string, error) {
	var op struct {
		Type string `json:"_type"`
	}
	if err := json.Unmarshal(raw, &op); err != nil {
		return "", fmt.Errorf("decoding operator %.60s: %w", string(raw), err)
	}
	return op.Type, nil
}

func decodeOps(raws []json.RawMessage) ([]string, error) {
	ops := make([]string, 0, len(raws))
	for _, r := range raws {
		op, err := decodeOp(r)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func decodeArg(raw json.RawMessage) (*Arg, error) {
	var aux struct {
		Arg        string          `json:"arg"`
		Annotation json.RawMessage `json:"annotation"`
		Lineno     int             `json:"lineno"`
	}
	if err := json.Unmarshal(raw, &aux); err != nil {
		return nil, err
	}
	ann, err := decodeOptional(aux.Annotation)
	if err != nil {
		return nil, err
	}
	return &Arg{Name: aux.Arg, Annotation: ann, Lineno: aux.Lineno}, nil
}

func decodeArgList(raws []json.RawMessage) ([]*Arg, error) {
	args := make([]*Arg, 0, len(raws))
	for _, r := range raws {
		a, err := decodeArg(r)
		if err != nil {
			return nil, err
		}
		args = append(args, a)
	}
	return args, nil
}

func decodeArguments(raw json.RawMessage) (*Arguments, error) {
	var aux struct {
		Posonlyargs []json.RawMessage `json:"posonlyargs"`
		Args        []json.RawMessage `json:"args"`
		Vararg      json.RawMessage   `json:"vararg"`
		Kwonlyargs  []json.RawMessage `json:"kwonlyargs"`
		KwDefaults  []json.RawMessage `json:"kw_defaults"`
		Kwarg       json.RawMessage   `json:"kwarg"`
		Defaults    []json.RawMessage `json:"defaults"`
	}
	if err := json.Unmarshal(raw, &aux); err != nil {
		return nil, err
	}
	var a Arguments
	var err error
	if a.PosOnly, err = decodeArgList(aux.Posonlyargs); err != nil {
		return nil, err
	}
	if a.Args, err = decodeArgList(aux.Args); err != nil {
		return nil, err
	}
	if a.Kwonly, err = decodeArgList(aux.Kwonlyargs); err != nil {
		return nil, err
	}
	if present(aux.Vararg) {
		if a.Vararg, err = decodeArg(aux.Vararg); err != nil {
			return nil, err
		}
	}
	if present(aux.Kwarg) {
		if a.Kwarg, err = decodeArg(aux.Kwarg); err != nil {
			return nil, err
		}
	}
	if a.Defaults, err = decodeList(aux.Defaults); err != nil {
		return nil, err
	}
	// kw_defaults has a null for each keyword-only parameter without a default.
	for _, r := range aux.KwDefaults {
		if !present(r) {
			continue
		}
		n, err := decodeNode(r)
		if err != nil {
			return nil, err
		}
		a.KwDefaults = append(a.KwDefaults, n)
	}
	return &a, nil
}

// decodeConstant uses the pytype tag from the dumper to pick a Go type.
// The raw JSON text is parsed directly, so 1 stays an int64 and 1.0 a float64.
func decodeConstant(pytype string, raw json.RawMessage, line int) (Node, error) {
	c := &Constant{PyType: pytype, Lineno: line}
	switch pytype {
	case "none":
	case "bool":
		var v bool
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, fmt.Errorf("line %d: bad bool constant: %w", line, err)
		}
		c.Value = v
	case "str":
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, fmt.Errorf("line %d: bad str constant: %w", line, err)
		}
		c.Value = v
	case "int":
		v, err := strconv.ParseInt(string(raw), 10, 64)
		if err != nil {
			// Python ints are unbounded, int64 is not.
			return &Unsupported{Type: "integer literal out of int64 range", Lineno: line}, nil
		}
		c.Value = v
	case "float":
		text := string(raw)
		if strings.HasPrefix(text, `"`) {
			// inf and nan arrive as strings, JSON has no spelling for them.
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return nil, fmt.Errorf("line %d: bad float constant: %w", line, err)
			}
			text = s
		}
		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: bad float constant %q: %w", line, text, err)
		}
		c.Value = v
	case "":
		return nil, fmt.Errorf("line %d: Constant has no pytype tag", line)
	default:
		// bytes, complex, ellipsis
		return &Unsupported{Type: pytype + " constant", Lineno: line}, nil
	}
	return c, nil
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

func present(raw json.RawMessage) bool {
	return len(raw) > 0 && string(raw) != "null"
}
