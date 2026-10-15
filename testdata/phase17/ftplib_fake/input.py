import ftplib
import io
import socket
import threading

FILES = {"a.txt": b"alpha\nbeta\n", "b.bin": bytes(range(10))}
stored = {}


def serve(ls):
    conn, _ = ls.accept()
    f = conn.makefile("rwb", buffering=0)
    f.write(b"220 fake ftp\r\n")
    data_ls = None
    while True:
        line = f.readline()
        if not line:
            break
        parts = line.decode().rstrip("\r\n").split(" ", 1)
        c = parts[0].upper()
        arg = parts[1] if len(parts) > 1 else ""
        if c == "USER":
            f.write(b"331 need password\r\n")
        elif c == "PASS":
            f.write(b"230 logged in\r\n")
        elif c == "SYST":
            f.write(b"215 UNIX Type: L8\r\n")
        elif c == "TYPE":
            f.write(b"200 type set\r\n")
        elif c == "PWD":
            f.write(b'257 "/" is cwd\r\n')
        elif c == "CWD":
            f.write(b"250 ok\r\n" if arg in ("sub", "/") else b"550 no such dir\r\n")
        elif c == "PASV":
            data_ls = socket.socket()
            data_ls.bind(("127.0.0.1", 0))
            data_ls.listen(1)
            p = data_ls.getsockname()[1]
            f.write(("227 Entering Passive Mode (127,0,0,1,%d,%d)\r\n" % (p >> 8, p & 255)).encode())
        elif c in ("LIST", "NLST", "RETR", "STOR"):
            if c == "RETR" and arg not in FILES:
                f.write(b"550 no such file\r\n")
                continue
            f.write(b"150 here it comes\r\n")
            d, _ = data_ls.accept()
            if c == "NLST":
                d.sendall("".join(n + "\r\n" for n in sorted(FILES)).encode())
            elif c == "LIST":
                d.sendall("".join("-rw-r--r-- 1 u g %d Jan 1 00:00 %s\r\n" % (len(FILES[n]), n) for n in sorted(FILES)).encode())
            elif c == "RETR":
                d.sendall(FILES[arg])
            else:
                buf = b""
                while True:
                    chunk = d.recv(100)
                    if not chunk:
                        break
                    buf += chunk
                stored[arg] = buf
            d.close()
            data_ls.close()
            f.write(b"226 transfer complete\r\n")
        elif c == "SIZE":
            f.write(("213 %d\r\n" % len(FILES[arg])).encode() if arg in FILES else b"550 nope\r\n")
        elif c == "QUIT":
            f.write(b"221 goodbye\r\n")
            break
        else:
            f.write(b"502 not implemented\r\n")
    f.close()
    conn.close()


ls = socket.socket()
ls.bind(("127.0.0.1", 0))
ls.listen(1)
port = ls.getsockname()[1]
t = threading.Thread(target=serve, args=(ls,))
t.start()

ftp = ftplib.FTP()
print(ftp.connect("127.0.0.1", port, timeout=5))
print(ftp.login("user", "pw"))
print(ftp.pwd())
print(ftp.cwd("sub"))
try:
    ftp.cwd("nope")
except ftplib.error_perm as e:
    print("error_perm", e)
print(ftp.nlst())
lines = []
print(ftp.retrlines("LIST", lines.append))
print(lines)
chunks = []
print(ftp.retrbinary("RETR a.txt", chunks.append))
print(b"".join(chunks))
print(ftp.size("b.bin"))
print(ftp.storbinary("STOR up.bin", io.BytesIO(b"x" * 250)))
try:
    ftp.retrbinary("RETR missing", chunks.append)
except ftplib.error_perm as e:
    print("error_perm", e)
print(ftp.quit())
t.join()
ls.close()
print({k: len(v) for k, v in stored.items()})
print(ftplib.parse227("227 Entering Passive Mode (10,0,0,1,4,1)"))
print(ftplib.parse257('257 "/home/me" created'))
