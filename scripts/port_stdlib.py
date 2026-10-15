"""Copies a module of CPython's Lib/ into stdlib/ without its type annotations
(and without comments and docstring formatting, since the source goes through
ast.unparse), for modules whose annotations name types py2go does not model.

    uv run --python 3.13 python scripts/port_stdlib.py SRC DST

The result is a starting point: what py2go cannot compile is then edited by hand.
"""
import ast
import sys


class Strip(ast.NodeTransformer):
    def visit_arg(self, node):
        node.annotation = None
        return node

    def _fn(self, node):
        node.returns = None
        self.generic_visit(node)
        return node

    visit_FunctionDef = _fn
    visit_AsyncFunctionDef = _fn

    def visit_AnnAssign(self, node):
        self.generic_visit(node)
        if node.value is None:
            return None
        return ast.copy_location(ast.Assign(targets=[node.target], value=node.value, lineno=node.lineno), node)

    def visit_ImportFrom(self, node):
        if node.module == "__future__":
            return None
        return node


src, dst = sys.argv[1:3]
tree = ast.parse(open(src).read())
tree = Strip().visit(tree)
ast.fix_missing_locations(tree)
open(dst, "w").write(ast.unparse(tree) + "\n")
