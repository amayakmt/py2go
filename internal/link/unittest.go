package link

import "github.com/amayakmt/py2go/internal/pyast"

// injectTestClasses gives unittest.main() and loader.loadTestsFromModule()
// in the entry module the test case classes of that module as the keyword
// _classes. A running program has no module objects to scan, so the
// linker, which sees the classes, passes them.
func injectTestClasses(mod *pyast.Module) {
	modNames := map[string]bool{}  // names bound to the unittest module
	fromNames := map[string]bool{} // TestCase and main imported by name
	mainNames := map[string]bool{}
	for _, st := range mod.Body {
		switch x := st.(type) {
		case *pyast.Import:
			for _, a := range x.Names {
				if a.Name == "unittest" {
					if a.AsName != "" {
						modNames[a.AsName] = true
					} else {
						modNames["unittest"] = true
					}
				}
			}
		case *pyast.ImportFrom:
			if x.Module == "unittest" && x.Level == 0 {
				for _, a := range x.Names {
					n := a.AsName
					if n == "" {
						n = a.Name
					}
					switch a.Name {
					case "TestCase", "IsolatedAsyncioTestCase":
						fromNames[n] = true
					case "main":
						mainNames[n] = true
					}
				}
			}
		}
	}
	if len(modNames) == 0 && len(fromNames) == 0 && len(mainNames) == 0 {
		return
	}
	isTestBase := func(b pyast.Node, known map[string]bool) bool {
		switch v := b.(type) {
		case *pyast.Name:
			return fromNames[v.ID] || known[v.ID]
		case *pyast.Attribute:
			if n, ok := v.Value.(*pyast.Name); ok {
				return modNames[n.ID] && (v.Attr == "TestCase" || v.Attr == "IsolatedAsyncioTestCase")
			}
		}
		return false
	}
	known := map[string]bool{}
	var order []string
	for changed := true; changed; {
		changed = false
		for _, st := range mod.Body {
			cd, ok := st.(*pyast.ClassDef)
			if !ok || known[cd.Name] {
				continue
			}
			for _, b := range cd.Bases {
				if isTestBase(b, known) {
					known[cd.Name] = true
					changed = true
					break
				}
			}
		}
	}
	for _, st := range mod.Body {
		if cd, ok := st.(*pyast.ClassDef); ok && known[cd.Name] {
			order = append(order, cd.Name)
		}
	}
	var visit func(n pyast.Node)
	visit = func(n pyast.Node) {
		if c, ok := n.(*pyast.Call); ok && wantsClasses(c, modNames, mainNames) {
			has := false
			for _, k := range c.Keywords {
				if k.Arg == "_classes" {
					has = true
				}
			}
			if !has {
				elts := make([]pyast.Node, len(order))
				for i, name := range order {
					elts[i] = &pyast.Name{ID: name, Lineno: c.Lineno}
				}
				c.Keywords = append(c.Keywords, &pyast.Keyword{Arg: "_classes", Value: &pyast.List{Elts: elts, Lineno: c.Lineno}, Lineno: c.Lineno})
				c.HasKeywords = true
			}
		}
		walkChildren(n, visit)
	}
	for _, st := range mod.Body {
		visit(st)
	}
}

func wantsClasses(c *pyast.Call, modNames, mainNames map[string]bool) bool {
	switch f := c.Func.(type) {
	case *pyast.Name:
		return mainNames[f.ID]
	case *pyast.Attribute:
		if f.Attr == "loadTestsFromModule" {
			return true
		}
		if n, ok := f.Value.(*pyast.Name); ok && modNames[n.ID] && f.Attr == "main" {
			return true
		}
	}
	return false
}

// rewriteTestDecorators turns @unittest.skip(...), @unittest.skipIf(...),
// @unittest.skipUnless(...) and @unittest.expectedFailure on test methods and
// classes into class attributes the test runner reads (_unittest_skip_<name>,
// _unittest_skip_class, _unittest_xfail_<name>): a method decorator cannot
// attach attributes to a method here.
func rewriteTestDecorators(mod *pyast.Module) {
	modNames := map[string]bool{}
	fromNames := map[string]string{}
	for _, st := range mod.Body {
		switch x := st.(type) {
		case *pyast.Import:
			for _, a := range x.Names {
				if a.Name == "unittest" {
					if a.AsName != "" {
						modNames[a.AsName] = true
					} else {
						modNames["unittest"] = true
					}
				}
			}
		case *pyast.ImportFrom:
			if x.Module == "unittest" && x.Level == 0 {
				for _, a := range x.Names {
					n := a.AsName
					if n == "" {
						n = a.Name
					}
					fromNames[n] = a.Name
				}
			}
		}
	}
	if len(modNames) == 0 && len(fromNames) == 0 {
		return
	}
	// kindOf names the unittest decorator d is, and its arguments.
	kindOf := func(d pyast.Node) (string, []pyast.Node, bool) {
		var fn pyast.Node = d
		var args []pyast.Node
		if c, ok := d.(*pyast.Call); ok {
			fn, args = c.Func, c.Args
		}
		name := ""
		switch f := fn.(type) {
		case *pyast.Attribute:
			if n, ok := f.Value.(*pyast.Name); ok && modNames[n.ID] {
				name = f.Attr
			}
		case *pyast.Name:
			name = fromNames[f.ID]
		}
		switch name {
		case "skip", "skipIf", "skipUnless", "expectedFailure":
			return name, args, true
		}
		return "", nil, false
	}
	none := func(line int) pyast.Node { return &pyast.Constant{PyType: "none", Value: nil, Lineno: line} }
	reasonOf := func(kind string, args []pyast.Node, line int) pyast.Node {
		switch kind {
		case "skip":
			if len(args) == 1 {
				return args[0]
			}
		case "skipIf":
			if len(args) == 2 {
				return &pyast.IfExp{Test: args[0], Body: args[1], Orelse: none(line), Lineno: line}
			}
		case "skipUnless":
			if len(args) == 2 {
				return &pyast.IfExp{Test: args[0], Body: none(line), Orelse: args[1], Lineno: line}
			}
		}
		return nil
	}
	assign := func(name string, value pyast.Node, line int) pyast.Node {
		return &pyast.Assign{Targets: []pyast.Node{&pyast.Name{ID: name, Lineno: line}}, Value: value, Lineno: line}
	}
	var visit func(n pyast.Node)
	visit = func(n pyast.Node) {
		if cd, ok := n.(*pyast.ClassDef); ok {
			var extra []pyast.Node
			var keep []pyast.Node
			for _, d := range cd.Decorators {
				kind, args, ok := kindOf(d)
				if !ok {
					keep = append(keep, d)
					continue
				}
				if r := reasonOf(kind, args, cd.Lineno); r != nil {
					extra = append(extra, assign("_unittest_skip_class", r, cd.Lineno))
				}
			}
			cd.Decorators = keep
			for _, st := range cd.Body {
				fd, ok := st.(*pyast.FunctionDef)
				if !ok {
					continue
				}
				var keepD []pyast.Node
				for _, d := range fd.Decorators {
					kind, args, ok := kindOf(d)
					if !ok {
						keepD = append(keepD, d)
						continue
					}
					if kind == "expectedFailure" {
						extra = append(extra, assign("_unittest_xfail_"+fd.Name, &pyast.Constant{PyType: "bool", Value: true, Lineno: fd.Lineno}, fd.Lineno))
					} else if r := reasonOf(kind, args, fd.Lineno); r != nil {
						extra = append(extra, assign("_unittest_skip_"+fd.Name, r, fd.Lineno))
					}
				}
				fd.Decorators = keepD
				fd.HasDecorators = len(keepD) > 0
			}
			cd.Body = append(cd.Body, extra...)
		}
		walkChildren(n, visit)
	}
	for _, st := range mod.Body {
		visit(st)
	}
}
