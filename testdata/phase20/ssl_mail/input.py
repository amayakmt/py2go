import os
import socket
import ssl
import tempfile
import threading

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

import ftplib
import poplib
import smtplib

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

listener = socket.socket()
listener.bind(("127.0.0.1", 0))
listener.listen()
port = listener.getsockname()[1]
log = []


def lines(f):
    while True:
        line = f.readline()
        if not line:
            return
        yield line.decode().rstrip("\r\n")


def smtp_session(conn):
    f = conn.makefile("rwb")

    def send(s):
        f.write(s.encode() + b"\r\n")
        f.flush()

    send("220 fake ESMTP")
    tls = False
    while True:
        raw = f.readline()
        if not raw:
            break
        line = raw.decode().rstrip("\r\n")
        log.append(("smtp", tls, line.split()[0] if line else ""))
        cmd = line.upper()
        if cmd.startswith("EHLO"):
            if tls:
                send("250-fake\r\n250 SIZE 1000")
            else:
                send("250-fake\r\n250 STARTTLS")
        elif cmd == "STARTTLS":
            send("220 go ahead")
            f.close()
            conn = server_ctx.wrap_socket(conn, server_side=True)
            f = conn.makefile("rwb")
            tls = True
        elif cmd.startswith("MAIL") or cmd.startswith("RCPT"):
            send("250 ok")
        elif cmd == "DATA":
            send("354 end with .")
            body = []
            for l in lines(f):
                if l == ".":
                    break
                body.append(l)
            log.append(("body", body[-1]))
            send("250 queued")
        elif cmd == "QUIT":
            send("221 bye")
            break
    f.close()
    conn.close()


def pop_session(conn):
    conn = server_ctx.wrap_socket(conn, server_side=True)
    f = conn.makefile("rwb")

    def send(s):
        f.write(s.encode() + b"\r\n")
        f.flush()

    send("+OK fake pop3")
    for line in lines(f):
        log.append(("pop", line.split()[0]))
        cmd = line.split()[0].upper()
        if cmd in ("USER", "PASS"):
            send("+OK")
        elif cmd == "STAT":
            send("+OK 2 320")
        elif cmd == "QUIT":
            send("+OK bye")
            break
    f.close()
    conn.close()


def serve():
    conn, _ = listener.accept()
    smtp_session(conn)
    conn, _ = listener.accept()
    pop_session(conn)


t = threading.Thread(target=serve)
t.start()

with smtplib.SMTP("localhost", port) as s:
    print(s.ehlo()[0], s.has_extn("starttls"))
    print(s.starttls(context=client_ctx))
    print(s.ehlo()[0], s.has_extn("size"), isinstance(s.sock, ssl.SSLSocket))
    print(s.sendmail("a@example.com", ["b@example.com"], "Subject: hi\r\n\r\nhello over tls"))

p = poplib.POP3_SSL("localhost", port, context=client_ctx)
print(p.getwelcome(), p.user("me"), p.pass_("pw"), p.stat())
print(p.quit())
t.join()
listener.close()
print(log)

ftps = ftplib.FTP_TLS(context=client_ctx)
print(type(ftps).__name__, ftps.context is client_ctx, ssl.SSLError in ftplib.all_errors, issubclass(smtplib.SMTP_SSL, smtplib.SMTP))
