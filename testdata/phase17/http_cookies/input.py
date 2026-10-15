from http import cookies

C = cookies.SimpleCookie()
C["fig"] = "newton"
C["sugar"] = "wafer"
print(C.output())
C["rocky"] = "road"
C["rocky"]["path"] = "/cookie"
C["rocky"]["httponly"] = True
C["rocky"]["max-age"] = 3600
print(C["rocky"].output())
print(C.output(attrs=["path"], header="Cookie:"))
print(repr(C))
print(str(C["fig"]))
print(C["fig"].value, C["fig"].key, C["fig"].coded_value)
print(sorted(C.keys()), len(C), "fig" in C)

D = cookies.SimpleCookie()
D.load("chips=ahoy; vienna=finger")
print(D.output())
D = cookies.SimpleCookie()
D.load('keebler="E=everybody; L=\\"Loves\\"; fudge=;";')
print(D)
print(D["keebler"].value)
E = cookies.SimpleCookie('a=1; Path=/x; b=2')
print(E.output())
print(E["a"]["path"], repr(E["b"].value))
M = cookies.Morsel()
M.set("k", "v w", '"v w"')
print(M.output(), M.key, M.value, M.coded_value)
print(M == M.copy(), M.isReservedKey("Path"), M.isReservedKey("zzz"))
try:
    M["bogus"] = "x"
except cookies.CookieError as e:
    print("CookieError", e)
try:
    C["bad key"] = "x"
except cookies.CookieError as e:
    print("CookieError", e)
print(C["fig"].js_output().strip().splitlines()[2].strip())
