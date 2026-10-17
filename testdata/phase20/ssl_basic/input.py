import http.client
import http.server
import os
import socket
import ssl
import tempfile
import threading
import urllib.request

# A self-signed certificate for localhost and 127.0.0.1, valid 2024 to 2124.
CERT = """-----BEGIN CERTIFICATE-----
MIIB0zCCAXigAwIBAgIDKioqMAoGCCqGSM49BAMCMDYxCzAJBgNVBAYTAkpQMRMw
EQYDVQQKEwpweTJnbyB0ZXN0MRIwEAYDVQQDEwlsb2NhbGhvc3QwIBcNMjQwMTAx
MDAwMDAwWhgPMjEyNDAxMDEwMDAwMDBaMDYxCzAJBgNVBAYTAkpQMRMwEQYDVQQK
EwpweTJnbyB0ZXN0MRIwEAYDVQQDEwlsb2NhbGhvc3QwWTATBgcqhkjOPQIBBggq
hkjOPQMBBwNCAAS04KrKRhIGMCsqNQklqO85dOgooV9NTc07R/uhyQ2Eot/6xwGh
wHuUzdE7Pi8NbH3xl+HKGPtOg+sROOSqKvajo3MwcTAOBgNVHQ8BAf8EBAMCAoQw
EwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU
2PUlMr2uIBFu3Pb7/vNNacZCmZMwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAAB
MAoGCCqGSM49BAMCA0kAMEYCIQCtQ+YIvc1yYHGanV7g2stSLeaBxjfvAO8m71Lt
xazjQwIhAOxe0YHJtiAj2dQmdXUgfoyDK5Hw5yORdrIobaSDcA9m
-----END CERTIFICATE-----
"""
KEY = """-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgFACnQ+WJcvqaX2/e
eficPx077TqPHWbzu4Ch1PPC+1mhRANCAAS04KrKRhIGMCsqNQklqO85dOgooV9N
Tc07R/uhyQ2Eot/6xwGhwHuUzdE7Pi8NbH3xl+HKGPtOg+sROOSqKvaj
-----END PRIVATE KEY-----
"""

d = tempfile.mkdtemp()
certfile = os.path.join(d, "cert.pem")
keyfile = os.path.join(d, "key.pem")
with open(certfile, "w") as f:
    f.write(CERT)
with open(keyfile, "w") as f:
    f.write(KEY)

server_ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
server_ctx.load_cert_chain(certfile, keyfile)
client_ctx = ssl.create_default_context(cafile=certfile)
print(client_ctx.verify_mode == ssl.CERT_REQUIRED, client_ctx.check_hostname, server_ctx.verify_mode == ssl.CERT_NONE)
print(client_ctx.protocol == ssl.PROTOCOL_TLS_CLIENT, ssl.HAS_TLSv1_3, ssl.TLSVersion.TLSv1_2 < ssl.TLSVersion.TLSv1_3)

listener = socket.socket()
listener.bind(("127.0.0.1", 0))
listener.listen()
port = listener.getsockname()[1]


def serve(n):
    for _ in range(n):
        conn, _ = listener.accept()
        try:
            with server_ctx.wrap_socket(conn, server_side=True) as s:
                data = s.recv(100)
                s.sendall(data.upper())
                f = s.makefile("rwb")
                line = f.readline()
                f.write(b"echo: " + line)
                f.flush()
                f.close()
        except (ssl.SSLError, OSError) as e:
            pass


t = threading.Thread(target=serve, args=(3,))
t.start()

with socket.create_connection(("127.0.0.1", port)) as raw:
    with client_ctx.wrap_socket(raw, server_hostname="localhost") as s:
        print(s.version() in ("TLSv1.2", "TLSv1.3"), s.server_hostname, s.server_side)
        cert = s.getpeercert()
        print(cert["subject"], cert["subjectAltName"], cert["serialNumber"], cert["notAfter"])
        print(len(s.getpeercert(binary_form=True)) > 100)
        s.sendall(b"hello")
        print(s.recv(100))
        s.sendall(b"a line\n")
        print(s.makefile("rb").readline())

for hostname, ctx in (("localhost", ssl.create_default_context()), ("example.com", client_ctx)):
    with socket.create_connection(("127.0.0.1", port)) as raw:
        try:
            ctx.wrap_socket(raw, server_hostname=hostname)
        except ssl.SSLCertVerificationError as e:
            print(type(e).__name__, e.reason, e.verify_code, e.verify_message, isinstance(e, ssl.SSLError), isinstance(e, ValueError))
t.join()
listener.close()

pem = ssl.DER_cert_to_PEM_cert(ssl.PEM_cert_to_DER_cert(CERT))
print(pem == CERT, pem.startswith(ssl.PEM_HEADER))


class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        body = ("secure " + self.path).encode()
        self.send_response(200)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        pass


httpd = http.server.HTTPServer(("127.0.0.1", 0), Handler)
httpd.socket = server_ctx.wrap_socket(httpd.socket, server_side=True)
hport = httpd.server_address[1]
ht = threading.Thread(target=httpd.serve_forever)
ht.start()
try:
    with urllib.request.urlopen(f"https://localhost:{hport}/a", context=client_ctx) as r:
        print(r.status, r.read())
    conn = http.client.HTTPSConnection("localhost", hport, context=client_ctx)
    conn.request("GET", "/b")
    resp = conn.getresponse()
    print(resp.status, resp.read())
    conn.close()
finally:
    httpd.shutdown()
    ht.join()
    httpd.server_close()
