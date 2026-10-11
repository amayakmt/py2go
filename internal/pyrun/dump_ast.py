import ast
import json
import sys

# Version guard. AST node shapes change between Python minor versions.
# If this runs under the wrong interpreter, the tree will be subtly different
# Exit if the Python version is not 3.13.
if sys.version_info[:2] != (3, 13):
    sys.exit(f"py2go requires Python 3.13, got {sys.version_info.major}.{sys.version_info.minor}")


def to_dict(node):
    """
    Recursively convert an AST node into plain dicts, lists, and values.

    ast.AST objects can't be serialized to JSON directly. This function
    turns the whole tree into something json.dump accepts.

    A value in the tree is one of three things:
    - A node (ast.AST), like FunctionDef or Call    -> becomes a dict
    - A list of nodes, like a function body         -> becomes a list
    - A plain value, like a string, int, or None    -> returned as is
    """

    # Case 1: a node.
    if isinstance(node, ast.AST):
        # "_type" records which kind of node this is ("FunctionDef", "Call", ...).
        # The Go decoder reads this first to decide which struct to build.
        d = {"_type": type(node).__name__}

        # Every node class declares its children in _fields.
        # For FunctionDef that's name, args, body, decorator_list, returns, ...
        # Recurse into each field, since a field may itself be a node or a list.
        # getattr(..., None) guards against a field that isn't set.
        for field in node._fields:
            d[field] = to_dict(getattr(node, field, None))

        # Position info is stored in attributes, not _fields, so it needs
        # separate handling. Only some nodes have it (Module doesn't).
        # lineno is what makes "unsupported: Match at line 12" possible.
        for attr in ("lineno", "col_offset"):
            if hasattr(node, attr):
                d[attr] = getattr(node, attr)
        return d

    # Case 2: a list, such as body containing several statements.
    # Convert each element the same way.
    if isinstance(node, list):
        return [to_dict(x) for x in node]

    # Case 3: a plain value (str, int, float, bool, None). JSON handles these.
    return node


def main():
    # Expect exactly one argument: the path to the .py file to parse.
    # sys.argv[0] is the script name, so a valid call has length 2.
    if len(sys.argv) != 2:
        sys.exit("usage: dump_ast.py <file.py>")

    with open(sys.argv[1], encoding="utf-8") as f:
        # ast.parse turns source text into a tree of node objects.
        # It checks syntax only. It runs nothing and checks no types.
        tree = ast.parse(f.read(), filename=sys.argv[1])

    # Write JSON to stdout. The Go program captures stdout and decodes it.
    # default=repr is a fallback for values JSON can't represent, such as
    # bytes, complex numbers, and Ellipsis. They become strings instead of
    # crashing the dump. Will be handled later on.
    json.dump(to_dict(tree), sys.stdout, indent=2, default=repr)
    print()


if __name__ == "__main__":
    main()