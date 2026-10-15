import http.client
import http.server
import json
import threading


class Handler(http.server.BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def log_message(self, format, *args):
        pass

    def do_GET(self):
        if self.path == "/json":
            body = json.dumps({"path": self.path, "ua": self.headers.get("User-Agent", "")}).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
        elif self.path == "/missing":
            self.send_error(404, "no such thing")
        elif self.path == "/redirect":
            self.send_response(302)
            self.send_header("Location", "/json")
            self.send_header("Content-Length", "0")
            self.end_headers()
        else:
            body = b"hello from " + self.path.encode()
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

    def do_POST(self):
        n = int(self.headers.get("Content-Length", 0))
        data = self.rfile.read(n)
        body = b"got " + data
        self.send_response(201, "Made")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("X-Echo", self.headers.get("X-Test", "-"))
        self.end_headers()
        self.wfile.write(body)

    def do_PUT(self):
        self.send_response(204)
        self.end_headers()


srv = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
host, port = srv.server_address
t = threading.Thread(target=srv.serve_forever, kwargs={"poll_interval": 0.05})
t.start()

c = http.client.HTTPConnection(host, port, timeout=5)
c.request("GET", "/hello")
r = c.getresponse()
print(r.status, r.reason, r.getheader("Content-Type"), r.read(), r.version, r.closed)
c.request("GET", "/json", headers={"User-Agent": "py2go-test"})
r = c.getresponse()
print(r.status, json.loads(r.read()), r.getheaders()[:2] and r.getheader("content-length"))
c.request("POST", "/post", body=b"payload", headers={"X-Test": "yes"})
r = c.getresponse()
print(r.status, r.reason, r.read(), r.getheader("X-Echo"), r.msg["Content-Length"])
c.request("GET", "/missing")
r = c.getresponse()
body = r.read()
print(r.status, r.reason, b"no such thing" in body, r.getheader("Content-Type"))
c.request("PUT", "/x", body="text")
r = c.getresponse()
print(r.status, r.reason, r.read(), r.length)
c.request("GET", "/redirect")
r = c.getresponse()
print(r.status, r.getheader("Location"))
r.read()
c.close()
c = http.client.HTTPConnection(host, port)
c.request("HEAD", "/hello")
r = c.getresponse()
print(r.status, r.read(), r.getheader("Content-Length"))
c.close()
srv.shutdown()
srv.server_close()
t.join()

try:
    http.client.HTTPConnection("127.0.0.1", 1, timeout=2).request("GET", "/")
except ConnectionRefusedError as e:
    print("refused")
print(http.client.responses[404], http.client.OK, http.client.HTTP_PORT, http.client.HTTPS_PORT)
try:
    http.client.HTTPConnection("host", "notaport")
except http.client.InvalidURL as e:
    print(e)
conn = http.client.HTTPConnection("example.com:8080")
print(conn.host, conn.port)
try:
    conn.putrequest("GET", "/x y")
except http.client.InvalidURL as e:
    print(e)
