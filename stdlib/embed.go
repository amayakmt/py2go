// Package stdlib holds the modules of the Python standard library that py2go
// ships, written in Python on top of the native functions of the runtime.
package stdlib

import "embed"

// Files is the tree of modules: math.py, collections/__init__.py, ...
//
//go:embed *.py */*.py */*/*.py
var Files embed.FS
