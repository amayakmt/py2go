import_re = None
def render(tpl, ctx):
    out, i = [], 0
    while i < len(tpl):
        if tpl.startswith("{{", i):
            j = tpl.index("}}", i)
            expr = tpl[i + 2:j].strip()
            name, _, fmt = expr.partition(":")
            val = ctx
            for part in name.strip().split("."):
                val = val[part] if isinstance(val, dict) else getattr(val, part)
            out.append(format(val, fmt) if fmt else str(val))
            i = j + 2
        else:
            out.append(tpl[i])
            i += 1
    return "".join(out)

class Item:
    def __init__(self, name, price):
        self.name, self.price = name, price

ctx = {"user": {"name": "Zoë", "age": 31}, "item": Item("Café au lait", 3.5), "n": 2 ** 70, "ratio": 0.12345}
tpl = "Hi {{ user.name }} ({{user.age}}): {{ item.name }} costs {{ item.price:8.2f}}|{{ n:,}}|{{ ratio:.1%}}|{{ n:#x}}|{{ user.name:*^9}}"
print(render(tpl, ctx))
print(render("no vars", ctx), render("{{n}}{{n}}", ctx)[:30])
for bad in ("{{ missing }}", "{{ user.zip }}", "{{ unclosed", "{{ n:z }}"):
    try:
        print(render(bad, ctx))
    except (KeyError, AttributeError, ValueError) as e:
        print(type(e).__name__, e)
print("{0:>{1}} {2!r} {0!s:.2}".format("abcdef", 8, "q'z"), "{:{fill}{align}{w}}".format("x", fill="-", align="^", w=7), "{a[0]}-{a[1]}-{b[k]}".format(a=[1, 2], b={"k": "v"}), "{:,}".format(1234567.891), "{:,.2f}".format(1234567.891), "{:08.3f}".format(-3.14159), "{:+d}|{: d}|{:x}|{:X}|{:o}|{:b}|{:c}".format(5, 5, 255, 255, 8, 5, 97))
