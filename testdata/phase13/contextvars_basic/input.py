import contextvars
import threading

var = contextvars.ContextVar("var", default="dflt")
nodefault = contextvars.ContextVar("nodefault")
print(var.name, var.get(), var.get("other"))
try:
    nodefault.get()
except LookupError:
    print("LookupError")
print(nodefault.get("fallback"))
token = var.set("one")
print(var.get(), token.old_value is contextvars.Token.MISSING, token.var is var)
token2 = var.set("two")
print(token2.old_value, var.get())
var.reset(token2)
print(var.get())
var.reset(token)
print(var.get())
try:
    var.reset(token)
except RuntimeError as e:
    print("RuntimeError", "already been used" in str(e))
other = contextvars.ContextVar("other")
try:
    other.reset(var.set("x"))
except ValueError as e:
    print("ValueError", "different ContextVar" in str(e))
var.set("main")
ctx = contextvars.copy_context()
print(ctx[var], var in ctx, len(ctx) >= 1, ctx.get(other, "none"))


def inner():
    var.set("inner")
    return var.get()


print(ctx.run(inner), var.get(), ctx[var])
try:
    ctx.run(lambda: ctx.run(inner))
except RuntimeError as e:
    print("RuntimeError", "already entered" in str(e))
c2 = ctx.copy()
c2.run(var.set, "copy")
print(c2[var], ctx[var])
seen = []


def thread_body():
    seen.append(var.get())
    var.set("thread")
    seen.append(var.get())


t = threading.Thread(target=thread_body)
t.start()
t.join()
print(seen, var.get())
t2 = threading.Thread(target=contextvars.copy_context().run, args=(thread_body,))
seen.clear()
t2.start()
t2.join()
print(seen, var.get())
print(sorted(v.name for v in ctx.keys()))
print(repr(var).startswith("<ContextVar name='var' default='dflt'"))
