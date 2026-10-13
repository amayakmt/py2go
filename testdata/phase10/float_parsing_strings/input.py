good = ["0", "-0", "+0.0", "1", "1.", ".5", "-.5", "1e3", "1E3", "1e+3", "1e-3", "1_000.5", "1_0e1_0", "  3.14  ", "\t2\n", "inf", "-inf", "+Infinity", "INFINITY", "nan", "-NaN", "0x1p3" if False else "8", "123456789012345678901234567890", "1e308", "1e309", "-1e309", "1e-400", "4.9e-324", "2.5e-324", "0.1", "0.30000000000000004", "9007199254740993", "1.7976931348623157e308", "1.7976931348623159e308"]
for s in good:
    try:
        v = float(s)
        print(repr(s), v, repr(v), v.is_integer() if v == v and abs(v) != float("inf") else None)
    except ValueError as e:
        print(repr(s), "ValueError", e)
bad = ["", " ", "1 2", "1..2", "1e", "e1", ".", "+", "-", "1_", "_1", "1__0", "0x10", "1,5", "one", "١٢٣", "1e1.5", "--1", "+-1", "infinit", "nann", "1.5f", "١.٥"]
for s in bad:
    try:
        print(repr(s), float(s))
    except ValueError as e:
        print(repr(s), "ValueError", e)
