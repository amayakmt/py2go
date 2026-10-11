// Package codegen turns a pyast tree into Go source.
package codegen

import (
	"fmt"
	"go/format"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// generator holds state that accumulates while walking the tree.
type generator struct {
	usesFmt bool // set when a print() is emitted, so we import "fmt"
}

// Generate converts a decoded Python module into formatted Go source.
func Generate(root pyast.Node) ([]byte, error) {
	mod, ok := root.(*pyast.Module)
	if !ok {
		return nil, fmt.Errorf("codegen: root must be Module, got %T", root)
	}

	g := &generator{}
	var funcs, body strings.Builder

	// Function definitions go to package level.
	// Everything else goes into main(), in source order.
	for _, stmt := range mod.Body {
		var err error
		if fn, ok := stmt.(*pyast.FunctionDef); ok {
			err = g.funcDef(&funcs, fn)
		} else {
			err = g.stmt(&body, stmt)
		}
		if err != nil {
			return nil, err
		}
	}

	var out strings.Builder
	out.WriteString("package main\n\n")
	if g.usesFmt {
		out.WriteString("import \"fmt\"\n\n")
	}
	out.WriteString(funcs.String())
	out.WriteString("func main() {\n")
	out.WriteString(body.String())
	out.WriteString("}\n")

	// go/format fixes indentation and fails on invalid syntax,
	// so a generator bug surfaces here instead of at go build.
	src, err := format.Source([]byte(out.String()))
	if err != nil {
		return nil, fmt.Errorf("codegen produced invalid Go: %w\n%s", err, out.String())
	}
	return src, nil
}

func (g *generator) funcDef(b *strings.Builder, fn *pyast.FunctionDef) error {
	if fn.Name == "main" {
		return unsupported("function named main (collides with generated main)", fn.Line())
	}
	if fn.HasParams {
		return unsupported("function parameters", fn.Line())
	}
	if fn.Returns == nil {
		return unsupported("function without return annotation", fn.Line())
	}
	// "-> None" arrives as a Constant holding nil.
	if ret, ok := fn.Returns.(*pyast.Constant); !ok || ret.Value != nil {
		return unsupported("return type other than None", fn.Line())
	}

	fmt.Fprintf(b, "func %s() {\n", fn.Name)
	for _, s := range fn.Body {
		if err := g.stmt(b, s); err != nil {
			return err
		}
	}
	b.WriteString("}\n\n")
	return nil
}

func (g *generator) stmt(b *strings.Builder, n pyast.Node) error {
	switch n := n.(type) {
	case *pyast.Expr:
		s, err := g.expr(n.Value)
		if err != nil {
			return err
		}
		// Go rejects a bare expression like "text" as a statement,
		// and go/format won't catch it. Python docstrings hit this.
		if _, ok := n.Value.(*pyast.Call); !ok {
			return unsupported("expression statement that is not a call", n.Line())
		}
		b.WriteString(s + "\n")
		return nil
	case *pyast.Unsupported:
		return unsupported(n.Type, n.Line())
	default:
		return unsupported(nodeName(n), n.Line())
	}
}

func (g *generator) expr(n pyast.Node) (string, error) {
	switch n := n.(type) {
	case *pyast.Call:
		name, ok := n.Func.(*pyast.Name)
		if !ok {
			return "", unsupported("call target that is not a plain name", n.Line())
		}
		if n.HasKeywords {
			return "", unsupported("keyword arguments", n.Line())
		}
		args := make([]string, 0, len(n.Args))
		for _, a := range n.Args {
			s, err := g.expr(a)
			if err != nil {
				return "", err
			}
			args = append(args, s)
		}
		callee := name.ID
		if callee == "print" {
			g.usesFmt = true
			callee = "fmt.Println"
		}
		return callee + "(" + strings.Join(args, ", ") + ")", nil

	case *pyast.Name:
		return n.ID, nil

	case *pyast.Constant:
		if s, ok := n.Value.(string); ok {
			return strconv.Quote(s), nil
		}
		return "", unsupported(fmt.Sprintf("constant of type %T", n.Value), n.Line())

	case *pyast.Unsupported:
		return "", unsupported(n.Type, n.Line())

	default:
		return "", unsupported(nodeName(n), n.Line())
	}
}

func unsupported(what string, line int) error {
	return fmt.Errorf("unsupported: %s at line %d", what, line)
}

// nodeName turns "*pyast.FunctionDef" into "FunctionDef" for error messages.
func nodeName(n pyast.Node) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", n), "*pyast.")
}
