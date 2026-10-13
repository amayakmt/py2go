import re
TOKEN = re.compile(r"""
    (?P<NUMBER>\d+(?:\.\d+)?)
  | (?P<NAME>[A-Za-z_]\w*)
  | (?P<OP>[-+*/()=])
  | (?P<SPACE>\s+)
""", re.VERBOSE)
def tokens(src):
    pos = 0
    while pos < len(src):
        m = TOKEN.match(src, pos)
        if not m:
            raise SyntaxError("bad char %r at %d" % (src[pos], pos))
        if m.lastgroup != "SPACE":
            yield (m.lastgroup, m.group())
        pos = m.end()
print(list(tokens("x = 3.14 * (y + 42)")))
try:
    list(tokens("a $ b"))
except SyntaxError as e:
    print(e)
log = "2024-01-15 ERROR [db] Connection failed (code=500)\n2024-01-16 INFO [web] Started in 35ms\n2024-01-17 ERROR [web] Timeout (code=504)"
pat = re.compile(r"^(?P<date>\d{4}-\d\d-\d\d) (?P<level>[A-Z]+) \[(?P<src>\w+)\] (?P<msg>.*?)(?: \(code=(?P<code>\d+)\))?$", re.M)
for m in pat.finditer(log):
    print(m["date"], m["level"], m["src"], m["msg"], m["code"], m.groupdict(default="-")["code"])
print(re.findall(r"[\w.+-]+@[\w-]+\.[\w.-]+", "mail bob@example.com or ann.lee+x@sub.site.org now"), re.findall(r"\b(?:\d{1,3}\.){3}\d{1,3}\b", "ip 192.168.0.1 and 10.0.0.255"), re.sub(r"(\d{3})(\d{3})(\d{4})", r"(\1) \2-\3", "5551234567"))
csv = 'a,"b,c",d,"e ""q"" f"'
print(re.findall(r'("(?:[^"]|"")*"|[^,]*)(?:,|$)', csv)[:4], re.split(r",(?=(?:[^\"]*\"[^\"]*\")*[^\"]*$)", csv))
camel = "someVariableNameHTTPServer"
print(re.sub(r"(?<!^)(?=[A-Z])", "_", camel).lower(), re.sub(r"([a-z])([A-Z])", r"\1 \2", camel), re.findall(r"[A-Z]?[a-z]+|[A-Z]+(?![a-z])", camel))
html = '<a href="x.html">One</a> <a class="c" href=\'y.html\'>Two</a>'
print(re.findall(r"<a\s[^>]*href=[\"']([^\"']+)[\"'][^>]*>(.*?)</a>", html), re.sub(r"<[^>]+>", "", html))
