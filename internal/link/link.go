// Package link turns a project of several Python files into one module.
//
// The code generator compiles one module at a time and infers types over all
// of it, so the modules of a project are joined before it sees them. Every
// import of a module of the project is replaced, where it stands, by the
// body of that module the first time it runs, as CPython executes it. The
// names a module binds at its top level are renamed with the module's path as
// a prefix ("pkg.mod" gives "pkg__mod_pm_name"), and the uses of an imported
// module's attributes ("mod.f", "from mod import f") are resolved to those
// names when the program is compiled. Because a module is inlined where it is
// first imported, a circular import sees the importer as CPython shows it: the
// names bound before the import statement, and nothing after.
//
// Imports inside a function or a class body run once, before the top-level
// statement that contains them, since the body is joined into the module and
// cannot run later.
package link

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/internal/pyrun"
)

// Marker separates the module prefix of a mangled name from the name the
// program wrote. The code generator and the runtime cut names at its last
// occurrence to show the original.
const Marker = "_pm_"

// Loader parses one Python file.
type Loader func(path string) (*pyast.Module, error)

// Load joins the project that starts at the file entry into one module.
func Load(entry string, load Loader) (*pyast.Module, error) {
	abs, err := filepath.Abs(entry)
	if err != nil {
		return nil, err
	}
	mod, err := load(abs)
	if err != nil {
		return nil, err
	}
	rewriteTestDecorators(mod)
	injectTestClasses(mod)
	l := &linker{root: filepath.Dir(abs), load: load, mods: map[string]*module{}, enums: map[string]*enumInfo{}}
	m := &module{name: "__main__", file: abs, ast: mod, entry: true}
	l.mods[m.name] = m
	l.begin(m)
	body, err := l.inlineBody(m)
	if err != nil {
		return nil, err
	}
	return &pyast.Module{Body: append(l.moduleRegistrations(), body...)}, nil
}

type module struct {
	name      string // dotted name, "__main__" for the entry
	file      string
	pkg       bool // a package (its file is __init__.py or it is a directory)
	namespace bool // a directory without __init__.py
	entry     bool
	builtins  bool // the module of the built-in names
	ast       *pyast.Module
	prefix    string
	bound     map[string]bool // names bound at the top level
	started   bool            // its body is, or has been, inlined
	finished  bool            // its body has run to the end
	defined   map[string]bool // names bound by the statements run so far
	aliases   map[string]alias
	allNames  []string // names exported by "from m import *"

	annotationMods map[string]bool // names bound to the typing or dataclasses module
}

// alias is what an imported name stands for in one module.
type alias struct {
	mod  *module // the module imported, or the module the name comes from
	name string  // "" when the alias is the module itself
}

type linker struct {
	root  string
	load  Loader
	mods  map[string]*module
	enums map[string]*enumInfo

	valueMods []*module // modules used as values, in order of first use
}

func (m *module) pkgName() string {
	if m.entry {
		return ""
	}
	if m.pkg {
		return m.name
	}
	if i := strings.LastIndex(m.name, "."); i >= 0 {
		return m.name[:i]
	}
	return ""
}

// begin prepares a module: its prefix and the names it binds.
func (l *linker) begin(m *module) {
	if !m.entry {
		m.prefix = strings.ReplaceAll(m.name, ".", "__") + Marker
	}
	m.bound = map[string]bool{}
	m.aliases = map[string]alias{}
	collectBound(m.ast.Body, m.bound)
	l.scanEnums(m)
	m.allNames = allList(m.ast.Body)
	for _, name := range globalNames(m.ast.Body) {
		m.bound[name] = true
	}
}

// find locates the module with the given dotted name on the project's path,
// then among the modules py2go ships.
func (l *linker) find(name string) (*module, error) {
	if m, ok := l.mods[name]; ok {
		return m, nil
	}
	if name == "builtins" {
		// The built-in names: nothing is renamed and nothing runs.
		m := &module{name: name, ast: &pyast.Module{}, builtins: true, started: true, finished: true, defined: map[string]bool{}}
		m.bound, m.aliases = map[string]bool{}, map[string]alias{}
		l.mods[name] = m
		return m, nil
	}
	parts := strings.Split(name, ".")
	root := l.root
	if !existsIn(root, parts[0]) {
		root = stdlibRoot()
		if root == "" || !existsIn(root, parts[0]) {
			return nil, nil
		}
	}
	dir := root
	for i, part := range parts {
		last := i == len(parts)-1
		sub := filepath.Join(dir, part)
		pkgName := strings.Join(parts[:i+1], ".")
		if _, ok := l.mods[pkgName]; !ok {
			var m *module
			switch {
			case fileExists(filepath.Join(sub, "__init__.py")):
				m = &module{name: pkgName, file: filepath.Join(sub, "__init__.py"), pkg: true}
			case fileExists(sub + ".py"):
				if !last {
					return nil, nil // a module is not a package
				}
				m = &module{name: pkgName, file: sub + ".py"}
			case dirExists(sub):
				m = &module{name: pkgName, file: sub, pkg: true, namespace: true}
			default:
				return nil, nil
			}
			if m.namespace {
				m.ast = &pyast.Module{}
			} else {
				ast, err := l.load(m.file)
				if err != nil {
					return nil, err
				}
				m.ast = ast
				rewriteTestDecorators(ast)
			}
			l.mods[pkgName] = m
			l.begin(m)
		}
		dir = sub
	}
	return l.mods[name], nil
}

func existsIn(dir, top string) bool {
	p := filepath.Join(dir, top)
	return fileExists(p+".py") || dirExists(p)
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// ---- inlining ----

// inlineBody is the module's statements, with its imports resolved and its
// names renamed.
func (l *linker) inlineBody(m *module) ([]pyast.Node, error) {
	m.started = true
	m.defined = map[string]bool{}
	r := &rewriter{l: l, m: m}
	// Imports inside functions and classes run before the statement that
	// holds them.
	var out []pyast.Node
	for _, st := range m.ast.Body {
		hoisted, err := r.hoist(st)
		if err != nil {
			return nil, err
		}
		out = append(out, hoisted...)
		res, err := r.stmt(st)
		if err != nil {
			return nil, err
		}
		out = append(out, res...)
		collectStmt(st, m.defined)
	}
	m.finished = true
	return out, nil
}

// importModule resolves a dotted name and the packages above it, and returns
// the statements that run them for the first time, in order.
func (l *linker) importModule(from *module, name string, line int) ([]pyast.Node, *module, error) {
	m, err := l.find(name)
	if err != nil {
		return nil, nil, err
	}
	if m == nil {
		return nil, nil, nil
	}
	var out []pyast.Node
	parts := strings.Split(name, ".")
	for i := range parts {
		pm, err := l.find(strings.Join(parts[:i+1], "."))
		if err != nil {
			return nil, nil, err
		}
		if pm.started {
			continue
		}
		body, err := l.inlineBody(pm)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, body...)
	}
	return out, m, nil
}

type rewriter struct {
	l      *linker
	m      *module
	scopes []*scope

	// Inside a classmethod of an enum, cls names the class.
	enumCls       string
	enumOf        *enumInfo
	enumClassBody *enumInfo // the enum whose body is being rewritten
}

type scope struct {
	names map[string]bool
	class bool
}

// block is several statements standing for one.
type block struct{ stmts []pyast.Node }

func (b *block) Line() int { return 0 }

// hoist inlines the modules imported inside a function or class body.
func (r *rewriter) hoist(st pyast.Node) ([]pyast.Node, error) {
	var out []pyast.Node
	var firstErr error
	var visit func(n pyast.Node, nested bool)
	visit = func(n pyast.Node, nested bool) {
		walkChildren(n, func(c pyast.Node) {
			switch x := c.(type) {
			case *pyast.FunctionDef, *pyast.ClassDef, *pyast.Lambda:
				visit(c, true)
			case *pyast.Import:
				if nested {
					res, err := r.importStmt(x)
					if err != nil && firstErr == nil {
						firstErr = err
					}
					out = append(out, res...)
				}
			case *pyast.ImportFrom:
				if nested {
					res, _, err := r.importFromStmt(x)
					if err != nil && firstErr == nil {
						firstErr = err
					}
					out = append(out, res...)
				}
			default:
				visit(c, nested)
			}
		})
	}
	switch st.(type) {
	case *pyast.Import, *pyast.ImportFrom:
		return nil, nil
	}
	// An import that is not a statement of the module itself (in a def, a
	// class or a compound statement) runs once, before its statement.
	visit(st, true)
	return out, firstErr
}

// importStmt handles "import a.b.c as d" for the modules of the project.
// Imports of anything else are left as they are.
func (r *rewriter) importStmt(n *pyast.Import) ([]pyast.Node, error) {
	var out []pyast.Node
	for _, a := range n.Names {
		body, m, err := r.l.importModule(r.m, a.Name, n.Lineno)
		if err != nil {
			return nil, err
		}
		if m == nil {
			return nil, nil
		}
		out = append(out, body...)
		if a.AsName != "" {
			r.m.aliases[a.AsName] = alias{mod: m}
		} else {
			top, _ := r.l.find(strings.Split(a.Name, ".")[0])
			r.m.aliases[strings.Split(a.Name, ".")[0]] = alias{mod: top}
		}
	}
	return out, nil
}

// importFromStmt handles "from m import a, b as c" and "from m import *".
func (r *rewriter) importFromStmt(n *pyast.ImportFrom) (body, raises []pyast.Node, err error) {
	name, ok := r.absolute(n)
	if !ok {
		return nil, nil, nil
	}
	body, m, err := r.l.importModule(r.m, name, n.Lineno)
	if err != nil {
		return nil, nil, err
	}
	if m == nil {
		return nil, nil, nil
	}
	out := body
	for _, a := range n.Names {
		if a.Name == "*" {
			for _, nm := range r.l.exports(m) {
				r.m.aliases[nm] = r.l.resolveName(m, nm)
			}
			continue
		}
		bind := a.AsName
		if bind == "" {
			bind = a.Name
		}
		if m.started && !m.finished && !m.defined[a.Name] && m.bound[a.Name] {
			raises = append(raises, raiseImportError(fmt.Sprintf("cannot import name '%s' from partially initialized module '%s' (most likely due to a circular import) (%s)", a.Name, m.name, m.file), n.Lineno))
			continue
		}
		if m.provides(r.l, a.Name) {
			r.m.aliases[bind] = r.l.resolveName(m, a.Name)
			continue
		}
		// A submodule of a package is an attribute of it once imported.
		sub, err := r.l.find(name + "." + a.Name)
		if err != nil {
			return nil, nil, err
		}
		if sub != nil && m.pkg {
			subBody, _, err := r.l.importModule(r.m, sub.name, n.Lineno)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, subBody...)
			r.m.aliases[bind] = alias{mod: sub}
			continue
		}
		raises = append(raises, raiseImportError(fmt.Sprintf("cannot import name '%s' from '%s' (%s)", a.Name, m.name, m.file), n.Lineno))
	}
	return out, raises, nil
}

// absolute is the dotted name a "from" import refers to.
func (r *rewriter) absolute(n *pyast.ImportFrom) (string, bool) {
	if n.Level == 0 {
		return n.Module, n.Module != ""
	}
	base := r.m.pkgName()
	if r.m.entry {
		return "", false
	}
	parts := []string{}
	if base != "" {
		parts = strings.Split(base, ".")
	}
	if n.Level-1 > len(parts) {
		return "", false
	}
	parts = parts[:len(parts)-(n.Level-1)]
	if n.Module != "" {
		parts = append(parts, strings.Split(n.Module, ".")...)
	}
	if len(parts) == 0 {
		return "", false
	}
	return strings.Join(parts, "."), true
}

// provides reports whether the module binds the name at its top level or
// imports it from another module of the project.
func (m *module) provides(l *linker, name string) bool {
	if m.builtins || m.bound[name] {
		return true
	}
	_, ok := m.aliases[name]
	return ok
}

// resolveName is what the name of a module stands for: the module's own
// binding, or what the module imported under that name.
func (l *linker) resolveName(m *module, name string) alias {
	if a, ok := m.aliases[name]; ok && !m.bound[name] {
		return a
	}
	return alias{mod: m, name: name}
}

// exports are the names "from m import *" takes.
func (l *linker) exports(m *module) []string {
	if m.allNames != nil {
		return m.allNames
	}
	var names []string
	for n := range m.bound {
		if !strings.HasPrefix(n, "_") {
			names = append(names, n)
		}
	}
	for n := range m.aliases {
		if !strings.HasPrefix(n, "_") && !m.bound[n] {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}

// allList reads __all__ = [...] from a module's statements.
func allList(body []pyast.Node) []string {
	var out []string
	found := false
	for _, st := range body {
		switch s := st.(type) {
		case *pyast.Assign:
			if len(s.Targets) == 1 && isName(s.Targets[0], "__all__") {
				out, found = strConsts(s.Value), true
			}
		case *pyast.AugAssign:
			if isName(s.Target, "__all__") && found {
				out = append(out, strConsts(s.Value)...)
			}
		}
	}
	if !found {
		return nil
	}
	return out
}

func strConsts(n pyast.Node) []string {
	var elts []pyast.Node
	switch x := n.(type) {
	case *pyast.List:
		elts = x.Elts
	case *pyast.Tuple:
		elts = x.Elts
	}
	out := []string{}
	for _, e := range elts {
		if c, ok := e.(*pyast.Constant); ok && c.PyType == "str" {
			out = append(out, c.Value.(string))
		}
	}
	return out
}

func isName(n pyast.Node, id string) bool {
	nm, ok := n.(*pyast.Name)
	return ok && nm.ID == id
}

// missingName is the first prefix of a dotted name that is not a module of
// the project.
func (l *linker) missingName(name string) string {
	parts := strings.Split(name, ".")
	for i := range parts {
		p := strings.Join(parts[:i+1], ".")
		if m, _ := l.find(p); m == nil {
			return p
		}
	}
	return name
}

func raiseNotFound(name string, line int) pyast.Node {
	return raiseClass("ModuleNotFoundError", "No module named '"+name+"'", line)
}

func raiseImportError(msg string, line int) pyast.Node {
	return raiseClass("ImportError", msg, line)
}

func raiseClass(class, msg string, line int) pyast.Node {
	return &pyast.Raise{
		Exc: &pyast.Call{
			Func:   &pyast.Name{ID: class, Lineno: line},
			Args:   []pyast.Node{&pyast.Constant{PyType: "str", Value: msg, Lineno: line}},
			Lineno: line,
		},
		Lineno: line,
	}
}

// ---- the rewrite ----

// stmt rewrites one statement and returns what replaces it.
func (r *rewriter) stmt(n pyast.Node) ([]pyast.Node, error) {
	switch x := n.(type) {
	case *pyast.Import:
		var rest []pyast.Alias
		for _, a := range x.Names {
			if a.Name == "typing" || a.Name == "dataclasses" {
				if r.m.annotationMods == nil {
					r.m.annotationMods = map[string]bool{}
				}
				bind := a.AsName
				if bind == "" {
					bind = a.Name
				}
				r.m.annotationMods[bind] = true
				continue
			}
			rest = append(rest, a)
		}
		if len(rest) != len(x.Names) {
			if r.m.annotationMods["dataclasses"] || len(r.m.annotationMods) > 0 {
				for _, a := range x.Names {
					if a.Name == "dataclasses" {
						imp := &pyast.ImportFrom{Module: "dataclasses", Names: []pyast.Alias{{Name: "dataclass"}, {Name: "field"}}, Lineno: x.Lineno}
						if len(rest) == 0 {
							return []pyast.Node{imp}, nil
						}
						x = &pyast.Import{Names: rest, Lineno: x.Lineno}
						return append([]pyast.Node{imp}, x), nil
					}
				}
			}
			if len(rest) == 0 {
				return passIfEmpty(nil, x.Lineno), nil
			}
			x = &pyast.Import{Names: rest, Lineno: x.Lineno}
			n = x
		}
		for _, a := range x.Names {
			if missing, gone := r.l.unavailable(a.Name); gone {
				return []pyast.Node{raiseNotFound(missing, x.Lineno)}, nil
			}
		}
		res, err := r.importStmt(x)
		if err != nil {
			return nil, err
		}
		if res == nil && !r.isProject(x) {
			for _, a := range x.Names {
				if top := strings.Split(a.Name, ".")[0]; !stdlibModules[top] {
					return []pyast.Node{raiseNotFound(r.l.missingName(a.Name), x.Lineno)}, nil
				}
			}
			return []pyast.Node{n}, nil
		}
		return passIfEmpty(res, x.Lineno), nil
	case *pyast.ImportFrom:
		if name, ok := r.absolute(x); ok {
			if missing, gone := r.l.unavailable(name); gone {
				return []pyast.Node{raiseNotFound(missing, x.Lineno)}, nil
			}
		}
		res, raises, err := r.importFromStmt(x)
		if err != nil {
			return nil, err
		}
		res = append(res, raises...)
		name, ok := r.absolute(x)
		if x.Level > 0 && !ok {
			return []pyast.Node{raiseImportError("attempted relative import with no known parent package", x.Lineno)}, nil
		}
		if !ok || !r.isProjectName(name) {
			if top := strings.Split(name, ".")[0]; ok && !stdlibModules[top] {
				return []pyast.Node{raiseNotFound(r.l.missingName(name), x.Lineno)}, nil
			}
			return []pyast.Node{n}, nil
		}
		return passIfEmpty(res, x.Lineno), nil
	}
	res, err := r.node(n)
	if err != nil {
		return nil, err
	}
	if b, ok := res.(*block); ok {
		return b.stmts, nil
	}
	return []pyast.Node{res}, nil
}

func passIfEmpty(stmts []pyast.Node, line int) []pyast.Node {
	if len(stmts) == 0 {
		return []pyast.Node{&pyast.Pass{Lineno: line}}
	}
	return stmts
}

func (r *rewriter) isProject(n *pyast.Import) bool {
	for _, a := range n.Names {
		if r.isProjectName(a.Name) {
			return true
		}
	}
	return false
}

func (r *rewriter) isProjectName(name string) bool {
	m, err := r.l.find(name)
	return err == nil && m != nil
}

// local reports whether a name is bound by an enclosing function, lambda,
// comprehension or (for code directly in it) class body.
func (r *rewriter) local(name string) bool {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		s := r.scopes[i]
		if s.class && i != len(r.scopes)-1 {
			continue
		}
		if s.names[name] {
			return true
		}
	}
	return false
}

func (r *rewriter) push(names map[string]bool, class bool) {
	r.scopes = append(r.scopes, &scope{names: names, class: class})
}

func (r *rewriter) pop() { r.scopes = r.scopes[:len(r.scopes)-1] }

// mangle is the name a module-level name has in the joined module.
func (r *rewriter) mangle(name string) string {
	if r.m.entry || r.local(name) || !r.m.bound[name] {
		return name
	}
	return r.m.prefix + name
}

// bindName renames the target of a def, class or except clause at the scope
// where it binds.
func (r *rewriter) bindName(name string) string {
	if len(r.scopes) == 0 {
		return r.mangle(name)
	}
	return name
}

func (r *rewriter) constantName(id string) pyast.Node {
	switch id {
	case "__name__":
		return &pyast.Constant{PyType: "str", Value: r.m.name}
	case "__package__":
		if r.m.entry {
			return &pyast.Constant{PyType: "none"}
		}
		return &pyast.Constant{PyType: "str", Value: r.m.pkgName()}
	}
	return nil
}

// node rewrites an expression or a compound statement.
func (r *rewriter) node(n pyast.Node) (pyast.Node, error) {
	switch x := n.(type) {
	case nil:
		return nil, nil
	case *pyast.Name:
		if r.local(x.ID) {
			return x, nil
		}
		if a, ok := r.m.aliases[x.ID]; ok {
			if a.name == "" {
				return r.l.moduleValue(a.mod, x.Lineno)
			}
			return &pyast.Name{ID: a.mod.prefixOf() + a.name, Lineno: x.Lineno}, nil
		}
		if c := r.constantName(x.ID); c != nil && !r.m.bound[x.ID] {
			return c, nil
		}
		x.ID = r.mangle(x.ID)
		return x, nil
	case *pyast.Attribute:
		if res, ok, err := r.attribute(x); ok || err != nil {
			if err == nil {
				res = r.enumUse(res)
			}
			return res, err
		}
	case *pyast.FunctionDef:
		return r.function(x)
	case *pyast.ClassDef:
		return r.class(x)
	case *pyast.Lambda:
		return r.lambda(x)
	case *pyast.ListComp:
		return x, r.comprehension(x.Generators, []*pyast.Node{&x.Elt})
	case *pyast.SetComp:
		return x, r.comprehension(x.Generators, []*pyast.Node{&x.Elt})
	case *pyast.GeneratorExp:
		return x, r.comprehension(x.Generators, []*pyast.Node{&x.Elt})
	case *pyast.DictComp:
		return x, r.comprehension(x.Generators, []*pyast.Node{&x.Key, &x.Value})
	case *pyast.Assign:
		if err := r.children(x); err != nil {
			return nil, err
		}
		if cls, ok, err := r.expandNamedTuple(x); ok || err != nil {
			return cls, err
		}
		// sys.stdout = f and sys.stderr = f tell the runtime where print writes.
		if len(x.Targets) == 1 {
			if nm, ok := x.Targets[0].(*pyast.Name); ok && (nm.ID == "sys_pm_stdout" || nm.ID == "sys_pm_stderr") {
				x.Value = &pyast.Call{Func: &pyast.Name{ID: "_native", Lineno: x.Lineno},
					Args: []pyast.Node{&pyast.Constant{PyType: "str", Value: "sys.set_" + strings.TrimPrefix(nm.ID, "sys_pm_"), Lineno: x.Lineno}, x.Value}, Lineno: x.Lineno}
			}
		}
		return x, nil
	case *pyast.Global:
		for i, nm := range x.Names {
			x.Names[i] = r.mangleGlobal(nm)
		}
		return x, nil
	case *pyast.ExceptHandler:
		if x.Name != "" {
			x.Name = r.bindName(x.Name)
		}
	case *pyast.MatchAs:
		if x.Name != "" {
			x.Name = r.bindName(x.Name)
		}
	case *pyast.MatchStar:
		if x.Name != "" {
			x.Name = r.bindName(x.Name)
		}
	case *pyast.MatchMapping:
		if x.Rest != "" {
			x.Rest = r.bindName(x.Rest)
		}
	}
	if err := r.children(n); err != nil {
		return nil, err
	}
	return r.enumUse(n), nil
}

func (r *rewriter) mangleGlobal(name string) string {
	if r.m.entry || !r.m.bound[name] {
		return name
	}
	return r.m.prefix + name
}

func (m *module) prefixOf() string {
	if m.entry || m.builtins {
		return ""
	}
	return m.prefix
}

// attribute resolves "m.f" and "pkg.sub.f" where m is a module of the
// project.
func (r *rewriter) attribute(x *pyast.Attribute) (pyast.Node, bool, error) {
	chain := []string{x.Attr}
	var base pyast.Node = x.Value
	for {
		a, ok := base.(*pyast.Attribute)
		if !ok {
			break
		}
		chain = append([]string{a.Attr}, chain...)
		base = a.Value
	}
	nm, ok := base.(*pyast.Name)
	if !ok || r.local(nm.ID) {
		return nil, false, nil
	}
	// typing.List, dataclasses.dataclass: the compiler knows these names
	// without an import, so the module qualifier is dropped.
	if len(chain) == 1 && r.m.annotationMods[nm.ID] {
		return &pyast.Name{ID: x.Attr, Lineno: x.Lineno}, true, nil
	}
	a, ok := r.m.aliases[nm.ID]
	if !ok {
		return nil, false, nil
	}
	cur := a
	for i, attr := range chain {
		if cur.name != "" && cur.mod.bound[cur.name+"__"+attr] {
			// chain.from_iterable: the module defines it as chain__from_iterable.
			cur = alias{mod: cur.mod, name: cur.name + "__" + attr}
			continue
		}
		if cur.name != "" {
			// A name of a module: the rest is attribute access on it.
			var node pyast.Node = &pyast.Name{ID: cur.mod.prefixOf() + cur.name, Lineno: x.Lineno}
			for _, rest := range chain[i:] {
				node = &pyast.Attribute{Value: node, Attr: rest, Lineno: x.Lineno}
			}
			return node, true, nil
		}
		m := cur.mod
		if i == len(chain)-1 && (attr == "__name__" || attr == "__package__") && !m.bound[attr] {
			if attr == "__name__" {
				return &pyast.Constant{PyType: "str", Value: m.name, Lineno: x.Lineno}, true, nil
			}
			return &pyast.Constant{PyType: "str", Value: m.pkgName(), Lineno: x.Lineno}, true, nil
		}
		if m.provides(r.l, attr) {
			cur = r.l.resolveName(m, attr)
			continue
		}
		if sub, _ := r.l.find(m.name + "." + attr); sub != nil && sub.started {
			cur = alias{mod: sub}
			continue
		}
		// Not there: reading the name raises AttributeError when it runs.
		cur = alias{mod: m, name: attr}
		continue
	}
	if cur.name == "" {
		v, err := r.l.moduleValue(cur.mod, x.Lineno)
		return v, true, err
	}
	return &pyast.Name{ID: cur.mod.prefixOf() + cur.name, Lineno: x.Lineno}, true, nil
}

// children rewrites the node's children in place.
func (r *rewriter) children(n pyast.Node) error {
	v := reflect.ValueOf(n)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return nil
	}
	return r.fields(v.Elem())
}

var nodeType = reflect.TypeOf((*pyast.Node)(nil)).Elem()

func (r *rewriter) fields(v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		if !v.Type().Field(i).IsExported() {
			continue
		}
		if err := r.value(v.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func (r *rewriter) value(f reflect.Value) error {
	switch f.Kind() {
	case reflect.Interface:
		if f.IsNil() || !f.Type().Implements(nodeType) {
			return nil
		}
		n, ok := f.Interface().(pyast.Node)
		if !ok {
			return nil
		}
		res, err := r.node(n)
		if err != nil {
			return err
		}
		if res != n && f.CanSet() {
			if res == nil {
				f.Set(reflect.Zero(f.Type()))
			} else {
				f.Set(reflect.ValueOf(res))
			}
		}
	case reflect.Pointer:
		if f.IsNil() {
			return nil
		}
		// A pointer to a node struct that is not held as a Node interface:
		// handlers, keywords, with items, match cases, arguments.
		switch x := f.Interface().(type) {
		case *pyast.ExceptHandler:
			_, err := r.node(x)
			return err
		case *pyast.Arguments:
			return nil
		}
		return r.fields(f.Elem())
	case reflect.Slice:
		if f.Type().Elem() == nodeType {
			out := make([]pyast.Node, 0, f.Len())
			for i := 0; i < f.Len(); i++ {
				n := f.Index(i).Interface()
				if n == nil {
					out = append(out, nil)
					continue
				}
				switch n.(type) {
				case *pyast.Import, *pyast.ImportFrom:
					res, err := r.stmt(n.(pyast.Node))
					if err != nil {
						return err
					}
					out = append(out, res...)
					continue
				}
				res, err := r.node(n.(pyast.Node))
				if err != nil {
					return err
				}
				if b, ok := res.(*block); ok {
					out = append(out, b.stmts...)
				} else {
					out = append(out, res)
				}
			}
			if f.CanSet() {
				f.Set(reflect.ValueOf(out))
			}
			return nil
		}
		for i := 0; i < f.Len(); i++ {
			if err := r.value(f.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Struct:
		return r.fields(f)
	}
	return nil
}

// function rewrites a def: decorators, defaults and annotations in the
// enclosing scope, the body in a new one.
func (r *rewriter) function(x *pyast.FunctionDef) (pyast.Node, error) {
	for i := range x.Decorators {
		n, err := r.node(x.Decorators[i])
		if err != nil {
			return nil, err
		}
		x.Decorators[i] = n
	}
	if err := r.outerOfArgs(x.Args); err != nil {
		return nil, err
	}
	if x.Returns != nil {
		n, err := r.node(x.Returns)
		if err != nil {
			return nil, err
		}
		x.Returns = n
	}
	x.Name = r.bindName(x.Name)
	names := paramNames(x.Args)
	collectBound(x.Body, names)
	for _, g := range globalNames(x.Body) {
		delete(names, g)
	}
	for _, g := range nonlocalNames(x.Body) {
		delete(names, g)
	}
	r.push(names, false)
	defer r.pop()
	if r.enumClassBody != nil && len(x.Args.Positional()) > 0 {
		for _, d := range x.Decorators {
			if nm, ok := d.(*pyast.Name); ok && nm.ID == "classmethod" {
				savedCls, savedOf := r.enumCls, r.enumOf
				r.enumCls, r.enumOf = x.Args.Positional()[0].Name, r.enumClassBody
				defer func() { r.enumCls, r.enumOf = savedCls, savedOf }()
			}
		}
	}
	if err := r.list(&x.Body); err != nil {
		return nil, err
	}
	// "mod.x = 1" became an assignment to a name of another module.
	assigned := map[string]bool{}
	collectBound(x.Body, assigned)
	var globals []string
	have := map[string]bool{}
	for _, g := range globalNames(x.Body) {
		have[g] = true
	}
	for nm := range assigned {
		if strings.Contains(nm, Marker) && !names[nm] && !have[nm] {
			globals = append(globals, nm)
		}
	}
	if len(globals) > 0 {
		sort.Strings(globals)
		x.Body = append([]pyast.Node{&pyast.Global{Names: globals, Lineno: x.Lineno}}, x.Body...)
	}
	return x, nil
}

// outerOfArgs rewrites the parts of a signature that run where the def does.
func (r *rewriter) outerOfArgs(a *pyast.Arguments) error {
	if a == nil {
		return nil
	}
	for i := range a.Defaults {
		n, err := r.node(a.Defaults[i])
		if err != nil {
			return err
		}
		a.Defaults[i] = n
	}
	for i := range a.KwDefaults {
		if a.KwDefaults[i] == nil {
			continue
		}
		n, err := r.node(a.KwDefaults[i])
		if err != nil {
			return err
		}
		a.KwDefaults[i] = n
	}
	for _, arg := range allArgs(a) {
		if arg.Annotation != nil {
			n, err := r.node(arg.Annotation)
			if err != nil {
				return err
			}
			arg.Annotation = n
		}
	}
	return nil
}

func allArgs(a *pyast.Arguments) []*pyast.Arg {
	out := append([]*pyast.Arg{}, a.PosOnly...)
	out = append(out, a.Args...)
	if a.Vararg != nil {
		out = append(out, a.Vararg)
	}
	out = append(out, a.Kwonly...)
	if a.Kwarg != nil {
		out = append(out, a.Kwarg)
	}
	return out
}

func paramNames(a *pyast.Arguments) map[string]bool {
	names := map[string]bool{}
	if a != nil {
		for _, arg := range allArgs(a) {
			names[arg.Name] = true
		}
	}
	return names
}

func (r *rewriter) list(body *[]pyast.Node) error {
	out := make([]pyast.Node, 0, len(*body))
	for _, st := range *body {
		hoisted, err := r.hoist(st)
		if err != nil {
			return err
		}
		out = append(out, hoisted...)
		res, err := r.stmt(st)
		if err != nil {
			return err
		}
		out = append(out, res...)
	}
	*body = out
	return nil
}

func (r *rewriter) class(x *pyast.ClassDef) (pyast.Node, error) {
	for i := range x.Decorators {
		n, err := r.node(x.Decorators[i])
		if err != nil {
			return nil, err
		}
		x.Decorators[i] = n
	}
	for i := range x.Bases {
		n, err := r.node(x.Bases[i])
		if err != nil {
			return nil, err
		}
		x.Bases[i] = n
	}
	for _, k := range x.Keywords {
		n, err := r.node(k.Value)
		if err != nil {
			return nil, err
		}
		k.Value = n
	}
	x.Name = r.bindName(x.Name)
	names := map[string]bool{}
	collectBound(x.Body, names)
	r.push(names, true)
	defer r.pop()
	savedBody := r.enumClassBody
	r.enumClassBody = r.l.enums[x.Name]
	err := r.list(&x.Body)
	r.enumClassBody = savedBody
	if err != nil {
		return nil, err
	}
	if err := expandTotalOrdering(x); err != nil {
		return nil, err
	}
	if info := r.l.enums[x.Name]; info != nil && len(r.scopes) == 1 {
		r.pop()
		defer r.push(names, true)
		return r.enumClass(x, info)
	}
	return x, nil
}

// totalOrdering gives, for the comparison a class defines, the bodies of the
// other three that functools.total_ordering would add.
var totalOrdering = map[string]map[string]string{
	"__lt__": {"__le__": "self < other or self == other", "__gt__": "not (self < other or self == other)", "__ge__": "not self < other"},
	"__le__": {"__lt__": "self <= other and self != other", "__gt__": "not self <= other", "__ge__": "not self <= other or self == other"},
	"__gt__": {"__lt__": "not (self > other or self == other)", "__le__": "not self > other", "__ge__": "self > other or self == other"},
	"__ge__": {"__lt__": "not self >= other", "__le__": "not self >= other or self == other", "__gt__": "self >= other and self != other"},
}

// expandTotalOrdering replaces @functools.total_ordering by the methods it
// writes, since a class decorator cannot run when the program is compiled.
func expandTotalOrdering(x *pyast.ClassDef) error {
	found := -1
	for i, d := range x.Decorators {
		if n, ok := d.(*pyast.Name); ok && (n.ID == "functools"+Marker+"total_ordering" || n.ID == "total_ordering") {
			found = i
		}
	}
	if found < 0 {
		return nil
	}
	x.Decorators = append(x.Decorators[:found], x.Decorators[found+1:]...)
	defined := map[string]bool{}
	for _, st := range x.Body {
		if fn, ok := st.(*pyast.FunctionDef); ok {
			defined[fn.Name] = true
		}
	}
	for _, root := range []string{"__lt__", "__le__", "__gt__", "__ge__"} {
		if !defined[root] {
			continue
		}
		var src strings.Builder
		src.WriteString("class _T:\n")
		for _, name := range []string{"__lt__", "__le__", "__gt__", "__ge__"} {
			if body, ok := totalOrdering[root][name]; ok && !defined[name] {
				fmt.Fprintf(&src, "    def %s(self, other):\n        return %s\n", name, body)
			}
		}
		mod, err := parseSnippet(src.String())
		if err != nil {
			return err
		}
		if cls, ok := mod.Body[0].(*pyast.ClassDef); ok {
			x.Body = append(x.Body, cls.Body...)
		}
		return nil
	}
	return fmt.Errorf("unsupported: total_ordering of a class that defines no comparison at line %d", x.Lineno)
}

func (r *rewriter) lambda(x *pyast.Lambda) (pyast.Node, error) {
	if err := r.outerOfArgs(x.Args); err != nil {
		return nil, err
	}
	r.push(paramNames(x.Args), false)
	defer r.pop()
	n, err := r.node(x.Body)
	if err != nil {
		return nil, err
	}
	x.Body = n
	return x, nil
}

// comprehension rewrites the clauses and the element expressions of a
// comprehension; its targets are local to it.
func (r *rewriter) comprehension(gens []*pyast.Comprehension, elts []*pyast.Node) error {
	names := map[string]bool{}
	for _, g := range gens {
		bindTargets(g.Target, names)
	}
	for i, g := range gens {
		if i == 0 {
			n, err := r.node(g.Iter)
			if err != nil {
				return err
			}
			g.Iter = r.enumIter(n)
			r.push(names, false)
			defer r.pop()
		} else {
			n, err := r.node(g.Iter)
			if err != nil {
				return err
			}
			g.Iter = r.enumIter(n)
		}
		t, err := r.node(g.Target)
		if err != nil {
			return err
		}
		g.Target = t
		for j := range g.Ifs {
			n, err := r.node(g.Ifs[j])
			if err != nil {
				return err
			}
			g.Ifs[j] = n
		}
	}
	for _, e := range elts {
		n, err := r.node(*e)
		if err != nil {
			return err
		}
		*e = n
	}
	return nil
}

// ---- names bound by statements ----

// collectBound adds the names the statements bind in their own scope: not
// those of nested functions, classes, lambdas or comprehensions, but the
// names of the defs and classes themselves.
func collectBound(body []pyast.Node, into map[string]bool) {
	for _, st := range body {
		collectStmt(st, into)
	}
}

func collectStmt(n pyast.Node, into map[string]bool) {
	switch x := n.(type) {
	case *pyast.FunctionDef:
		into[x.Name] = true
		return
	case *pyast.ClassDef:
		into[x.Name] = true
		return
	case *pyast.Assign:
		for _, t := range x.Targets {
			bindTargets(t, into)
		}
	case *pyast.AugAssign:
		bindTargets(x.Target, into)
	case *pyast.AnnAssign:
		bindTargets(x.Target, into)
	case *pyast.For:
		bindTargets(x.Target, into)
	case *pyast.With:
		for _, it := range x.Items {
			if it.Vars != nil {
				bindTargets(it.Vars, into)
			}
		}
	case *pyast.Try:
		for _, h := range x.Handlers {
			if h.Name != "" {
				into[h.Name] = true
			}
		}
	case *pyast.Match:
		for _, c := range x.Cases {
			bindPattern(c.Pattern, into)
		}
	case *pyast.TypeAlias:
		into[x.Name] = true
	}
	// Walrus targets and the nested statements of compound statements.
	walkChildren(n, func(c pyast.Node) {
		switch y := c.(type) {
		case *pyast.FunctionDef, *pyast.ClassDef, *pyast.Lambda:
			if f, ok := y.(*pyast.FunctionDef); ok {
				into[f.Name] = true
			}
			if cl, ok := y.(*pyast.ClassDef); ok {
				into[cl.Name] = true
			}
		case *pyast.ListComp, *pyast.SetComp, *pyast.DictComp, *pyast.GeneratorExp:
			walkChildren(y, func(cc pyast.Node) {
				if ne, ok := cc.(*pyast.NamedExpr); ok {
					bindTargets(ne.Target, into)
				}
			})
		case *pyast.NamedExpr:
			bindTargets(y.Target, into)
			collectStmt(y.Value, into)
		default:
			collectStmt(c, into)
		}
	})
}

func bindTargets(n pyast.Node, into map[string]bool) {
	switch x := n.(type) {
	case *pyast.Name:
		into[x.ID] = true
	case *pyast.Tuple:
		for _, e := range x.Elts {
			bindTargets(e, into)
		}
	case *pyast.List:
		for _, e := range x.Elts {
			bindTargets(e, into)
		}
	case *pyast.Starred:
		bindTargets(x.Value, into)
	}
}

func bindPattern(n pyast.Node, into map[string]bool) {
	switch x := n.(type) {
	case *pyast.MatchAs:
		if x.Name != "" {
			into[x.Name] = true
		}
		if x.Pattern != nil {
			bindPattern(x.Pattern, into)
		}
	case *pyast.MatchStar:
		if x.Name != "" {
			into[x.Name] = true
		}
	case *pyast.MatchSequence:
		for _, p := range x.Patterns {
			bindPattern(p, into)
		}
	case *pyast.MatchMapping:
		for _, p := range x.Patterns {
			bindPattern(p, into)
		}
		if x.Rest != "" {
			into[x.Rest] = true
		}
	case *pyast.MatchClass:
		for _, p := range x.Patterns {
			bindPattern(p, into)
		}
		for _, p := range x.KwdPatterns {
			bindPattern(p, into)
		}
	case *pyast.MatchOr:
		for _, p := range x.Patterns {
			bindPattern(p, into)
		}
	}
}

// globalNames are the names declared global anywhere under the statements.
func globalNames(body []pyast.Node) []string {
	return declared(body, func(n pyast.Node) []string {
		if g, ok := n.(*pyast.Global); ok {
			return g.Names
		}
		return nil
	})
}

func nonlocalNames(body []pyast.Node) []string {
	return declared(body, func(n pyast.Node) []string {
		if g, ok := n.(*pyast.Nonlocal); ok {
			return g.Names
		}
		return nil
	})
}

func declared(body []pyast.Node, pick func(pyast.Node) []string) []string {
	var out []string
	var visit func(n pyast.Node, depth int)
	visit = func(n pyast.Node, depth int) {
		if names := pick(n); names != nil {
			out = append(out, names...)
		}
		walkChildren(n, func(c pyast.Node) { visit(c, depth+1) })
	}
	for _, st := range body {
		visit(st, 0)
	}
	return out
}

// walkChildren calls f on each direct child node of n.
func walkChildren(n pyast.Node, f func(pyast.Node)) {
	var visit func(v reflect.Value)
	visit = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Interface:
			if !v.IsNil() {
				visit(v.Elem())
			}
		case reflect.Pointer:
			if v.IsNil() {
				return
			}
			if v.Type().Implements(nodeType) {
				f(v.Interface().(pyast.Node))
				return
			}
			visit(v.Elem())
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Type().Field(i).IsExported() {
					visit(v.Field(i))
				}
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				visit(v.Index(i))
			}
		}
	}
	v := reflect.ValueOf(n)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		visit(v.Elem())
	}
}

// LoadFile parses one file with the Python of the project and joins its
// imports.
func LoadFile(entry string) (*pyast.Module, error) {
	return Load(entry, parseFile)
}

func parseFile(path string) (*pyast.Module, error) {
	js, err := pyrun.DumpAST(path)
	if err != nil {
		return nil, err
	}
	root, err := pyast.Decode(js)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	mod, ok := root.(*pyast.Module)
	if !ok {
		return nil, fmt.Errorf("%s: root is %T, not a module", path, root)
	}
	return mod, nil
}
