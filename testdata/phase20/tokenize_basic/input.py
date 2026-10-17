import io
import token
import tokenize

src = '''def f(a, b=2):
    """doc"""
    return a + b  # sum

x = f"{f(1):>{4}} {{literal}}"
if x:
\tpass
'''
for t in tokenize.generate_tokens(io.StringIO(src).readline):
    print(token.tok_name[t.type], repr(t.string), t.start, t.end, t.exact_type == t.type)

data = src.encode()
toks = list(tokenize.tokenize(io.BytesIO(data).readline))
print(toks[0].type == tokenize.ENCODING, toks[0].string, len(toks))
print(tokenize.untokenize(toks).decode() == src)
print(tokenize.detect_encoding(io.BytesIO(b"# -*- coding: latin-1 -*-\nx = 1\n").readline))
print([tokenize.tok_name[t.exact_type] for t in tokenize.generate_tokens(io.StringIO("a += [1] ** 2\n").readline)][:6])

for bad in ['x = "abc\n', 'x = (1,\n', 'if x:\n  y\n z\n', 'x = """\n']:
    try:
        list(tokenize.generate_tokens(io.StringIO(bad).readline))
    except tokenize.TokenError as e:
        print("TokenError", e.args)
    except IndentationError as e:
        print(type(e).__name__, e.msg, e.lineno)
