import tomllib

doc = """
# a comment
title = "TOML Example"
count = 42
pi = 3.14159
neg = -17
hex = 0xDEADBEEF
big = 1_000_000
flag = true
off = false
inf_v = inf
nan_v = nan
when = 1979-05-27T07:32:00Z
local = 1979-05-27T07:32:00
day = 1979-05-27
clock = 07:32:00.5
list = [1, 2, 3]
mixed = [ "a", 1, [2, 3], {x = 1} ]
inline = { name = "tom", age = 30 }
multi = '''
line one
line two'''
basic = "tab\\there \\u00e9 \\"quoted\\""
literal = 'C:\\no\\escape'

[owner]
name = "Tom"
dob = 1979-05-27T07:32:00-08:00

[database]
ports = [ 8000, 8001 ]
enabled = true

[database.pool]
size = 10

[[servers]]
name = "alpha"
ip = "10.0.0.1"

[[servers]]
name = "beta"

[servers.extra]
x = 1

[a.b.c]
d = "deep"
"quoted key" = 1
"""
data = tomllib.loads(doc)
for k in data:
    print(k, repr(data[k]))
print(tomllib.loads("a = 1\nb.c = 2\nb.d = 3"))
print(tomllib.loads(""))
for bad in ["a = ", "a = 1\na = 2", "[x]\n[x]", "a = 'unterminated", "a = [1, 'x'", "= 1", "a = 1 b = 2", "a = 0x", "a = 2024-13-01", "[a]\nb = 1\n[a.b]\nc = 2"]:
    try:
        tomllib.loads(bad)
    except tomllib.TOMLDecodeError as e:
        print(type(e).__name__, e)
import io
print(tomllib.load(io.BytesIO(b'k = "v"')))
print(tomllib.loads("x = 1.5", parse_float=lambda s: s + "!"))
