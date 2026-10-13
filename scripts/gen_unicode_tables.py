#!/usr/bin/env python3
"""Regenerates runtime/strcase_tables.go and runtime/unicode_tables.go from
CPython 3.13's own str methods, so that the transpiled program classifies and
converts characters exactly as CPython does (Unicode 15.1.0).

    uv run --python 3.13 python scripts/gen_unicode_tables.py

The output is formatted like gofmt's, so `gofmt -l runtime` stays quiet.
"""
import sys
import unicodedata
from pathlib import Path

RUNTIME = Path(__file__).resolve().parent.parent / "runtime"
CODE_POINTS = [c for c in range(0x110000) if not 0xD800 <= c <= 0xDFFF]


def lit(s):
    return "".join("\\u%04x" % ord(ch) if ord(ch) <= 0xFFFF else "\\U%08x" % ord(ch) for ch in s)


def string_map(name, comment, pairs):
    out = [f"var {name} = map[rune]string{{"]
    out += [f'\t0x{c:04x}: "{lit(s)}",' for c, s in pairs]
    out.append("}")
    return "\n".join(out) if not comment else f"{comment}\n" + "\n".join(out)


def ranges(pred):
    out, start, prev = [], None, None
    for c in CODE_POINTS:
        if pred(c):
            if start is None:
                start = c
            elif c != prev + 1:
                out.append((start, prev))
                start = c
            prev = c
        elif start is not None:
            out.append((start, prev))
            start = None
    if start is not None:
        out.append((start, prev))
    return out


def range_table(name, pred):
    items = [f"{{0x{a:x}, 0x{b:x}}}" for a, b in ranges(pred)]
    lines = ["\t" + ", ".join(items[i:i + 6]) + "," for i in range(0, len(items), 6)]
    return f"var {name} = rangeTable{{\n" + "\n".join(lines) + "\n}"


def strcase():
    multi = lambda f: [(c, f(chr(c))) for c in CODE_POINTS if len(f(chr(c))) > 1]
    simple = [(c, ord(chr(c).casefold())) for c in CODE_POINTS
              if len(chr(c).casefold()) == 1 and chr(c).casefold() != chr(c).lower()]
    parts = [
        "package runtime",
        "// Code generated from CPython 3.13's str.upper, title, lower and casefold (Unicode 15.1.0). DO NOT EDIT.",
        string_map("specialUpper", "", multi(str.upper)),
        string_map("specialTitle", "", multi(str.title)),
        string_map("specialLower", "", multi(str.lower)),
        string_map("specialFold", "", multi(str.casefold)),
        "// simpleFold lists the characters whose casefold is one character other than their lower case.\n"
        "var simpleFold = map[rune]rune{\n" + "\n".join(f"\t0x{c:04x}: 0x{f:04x}," for c, f in simple) + "\n}",
    ]
    return "\n\n".join(parts) + "\n"


def tables():
    isident = lambda c: ("a" + chr(c)).isidentifier()
    parts = [
        "package runtime",
        "// Code generated from CPython 3.13's str predicates (Unicode %s). DO NOT EDIT." % unicodedata.unidata_version,
        "type rangeTable [][2]rune",
        range_table("decimalRanges", lambda c: chr(c).isdecimal()),
        range_table("digitRanges", lambda c: chr(c).isdigit()),
        range_table("numericRanges", lambda c: chr(c).isnumeric()),
        range_table("alphaRanges", lambda c: chr(c).isalpha()),
        range_table("spaceRanges", lambda c: chr(c).isspace()),
        range_table("printableRanges", lambda c: chr(c).isprintable()),
        range_table("identStartRanges", lambda c: chr(c).isidentifier()),
        range_table("identContinueRanges", isident),
        "// has reports whether r is in the table.\n"
        "func (t rangeTable) has(r rune) bool {\n"
        "\tlo, hi := 0, len(t)\n"
        "\tfor lo < hi {\n"
        "\t\tmid := (lo + hi) / 2\n"
        "\t\tswitch {\n"
        "\t\tcase r < t[mid][0]:\n"
        "\t\t\thi = mid\n"
        "\t\tcase r > t[mid][1]:\n"
        "\t\t\tlo = mid + 1\n"
        "\t\tdefault:\n"
        "\t\t\treturn true\n"
        "\t\t}\n"
        "\t}\n"
        "\treturn false\n"
        "}",
    ]
    return "\n\n".join(parts) + "\n"


if __name__ == "__main__":
    out = Path(sys.argv[1]) if len(sys.argv) > 1 else RUNTIME
    (out / "strcase_tables.go").write_text(strcase())
    (out / "unicode_tables.go").write_text(tables())
