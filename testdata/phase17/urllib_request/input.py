import http.server
import os
import tempfile
import threading
import urllib.error
import urllib.request


class H(http.server.BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        pass

    def do_GET(self):
        if self.path == "/old":
            self.send_response(301)
            self.send_header("Location", "/new")
            self.send_header("Content-Length", "0")
            self.end_headers()
        elif self.path == "/setck":
            self.send_response(200)
            self.send_header("Set-Cookie", "tok=xyz; Path=/")
            self.send_header("Content-Length", "2")
            self.end_headers()
            self.wfile.write(b"ok")
        elif self.path == "/getck":
            body = self.headers.get("Cookie", "none").encode()
            self.send_response(200)
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
        elif self.path == "/gone":
            self.send_error(404, "Gone fishing")
        else:
            body = ("path=" + self.path + " ua=" + self.headers.get("User-Agent", "")).encode()
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

    def do_POST(self):
        n = int(self.headers.get("Content-Length", 0))
        data = self.rfile.read(n)
        body = b"posted:" + data + b" " + self.headers.get("Content-Type", "").encode()
        self.send_response(200)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


srv = http.server.ThreadingHTTPServer(("127.0.0.1", 0), H)
port = srv.server_address[1]
t = threading.Thread(target=srv.serve_forever, kwargs={"poll_interval": 0.05})
t.start()
base = "http://127.0.0.1:%d" % port

with urllib.request.urlopen(base + "/a?x=1") as r:
    print(r.status, r.reason, r.getcode(), r.headers["Content-Type"], r.read())
    print(r.geturl() == base + "/a?x=1")
r = urllib.request.urlopen(base + "/old")
print(r.status, r.geturl() == base + "/new", r.read())
req = urllib.request.Request(base + "/p", data=b"k=v", headers={"x-thing": "1"})
print(req.get_method(), req.full_url == base + "/p", req.host == "127.0.0.1:%d" % port, req.selector, req.has_header("X-thing"))
print(urllib.request.urlopen(req).read())
try:
    urllib.request.urlopen(base + "/gone")
except urllib.error.HTTPError as e:
    print(e.code, e.reason, str(e), isinstance(e, urllib.error.URLError), e.headers["Content-Type"])
    print(b"Gone fishing" in e.read())
import http.cookiejar
jar = http.cookiejar.CookieJar()
op = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar))
print(op.open(base + "/setck").read(), len(jar))
print(op.open(base + "/getck").read())
try:
    urllib.request.urlopen("http://127.0.0.1:1/")
except urllib.error.URLError as e:
    print("urlerror", type(e.reason).__name__)

srv.shutdown()
srv.server_close()
t.join()

with urllib.request.urlopen("data:text/plain;base64,aGVsbG8=") as r:
    print(r.read(), r.headers["Content-type"])
d = tempfile.mkdtemp()
p = os.path.join(d, "f.txt")
with open(p, "w") as f:
    f.write("file body\n")
with urllib.request.urlopen("file://" + p) as r:
    print(r.read(), r.headers["Content-length"])
try:
    urllib.request.urlopen("ftp2://x/")
except urllib.error.URLError as e:
    print(e)
print(urllib.request.pathname2url("/a b/c"), urllib.request.url2pathname("/a%20b"))
os.remove(p)
os.rmdir(d)
