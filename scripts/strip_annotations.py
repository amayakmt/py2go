"""Writes a copy of the differential tests of phases 2 to 8 with every type
annotation removed, for the phase 9 exit test: the transpiler must give the
same output for the programs without their hints.

    python scripts/strip_annotations.py OUTDIR

Each case becomes OUTDIR/p<phase>_<name>/ with the stripped input.py and the
expected.txt / expected.err of the original (CPython does not look at
annotations, so its output is the same).

A program is left out when its annotations are its meaning: dataclass fields,
TypeVar and Generic classes, ClassVar. Those are the declarations phase 9 does
not replace.
"""

import ast
import glob
import os
import shutil
import sys

ANNOTATION_AS_MEANING = ("dataclass", "TypeVar", "ClassVar", "Generic[", "def __exit__")


class Strip(ast.NodeTransformer):
    def visit_arg(self, node):
        node.annotation = None
        return node

    def _function(self, node):
        node.returns = None
        self.generic_visit(node)
        return node

    visit_FunctionDef = _function
    visit_AsyncFunctionDef = _function

    def visit_AnnAssign(self, node):
        self.generic_visit(node)
        if node.value is None:
            return None
        return ast.copy_location(ast.Assign(targets=[node.target], value=node.value, lineno=node.lineno), node)


def main():
    out = sys.argv[1]
    root = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "testdata")
    shutil.rmtree(out, ignore_errors=True)
    written = left_out = 0
    for phase in range(2, 9):
        for d in sorted(glob.glob(os.path.join(root, f"phase{phase}", "*", ""))):
            with open(os.path.join(d, "input.py")) as f:
                src = f.read()
            if any(word in src for word in ANNOTATION_AS_MEANING):
                left_out += 1
                continue
            tree = ast.fix_missing_locations(Strip().visit(ast.parse(src)))
            name = f"p{phase}_{os.path.basename(d.rstrip(os.sep))}"
            case = os.path.join(out, name)
            os.makedirs(case)
            with open(os.path.join(case, "input.py"), "w") as f:
                f.write(ast.unparse(tree) + "\n")
            for expected in ("expected.txt", "expected.err"):
                if os.path.exists(os.path.join(d, expected)):
                    shutil.copy(os.path.join(d, expected), os.path.join(case, expected))
            written += 1
    print(f"{written} cases written, {left_out} left out", file=sys.stderr)


main()
