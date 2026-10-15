import socket
import threading
import errno

# TCP echo
srv = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
srv.bind(("127.0.0.1", 0))
srv.listen(5)
host, port = srv.getsockname()
print(host, port > 0, srv.family, srv.type, srv.proto, srv.getsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR) != 0)

def echo():
    conn, addr = srv.accept()
    with conn:
        print("accepted", addr[0], conn.getpeername() == addr)
        while True:
            data = conn.recv(1024)
            if not data:
                break
            conn.sendall(data.upper())

t = threading.Thread(target=echo)
t.start()
c = socket.create_connection((host, port))
c.sendall(b"hello ")
print(c.recv(100))
c.sendall(b"world")
print(c.recv(100), c.getsockname()[0], c.getpeername() == (host, port))
c.shutdown(socket.SHUT_WR)
print(c.recv(10))
c.close()
t.join()
srv.close()
print(c.fileno(), repr(c))
try:
    c.send(b"x")
except OSError as e:
    print(type(e).__name__, e.errno)

# UDP
a = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
a.bind(("127.0.0.1", 0))
b = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
b.sendto(b"datagram", a.getsockname())
data, addr = a.recvfrom(100)
print(data, addr[0], addr[1] == b.getsockname()[1])
a.settimeout(0.1)
try:
    a.recvfrom(10)
except socket.timeout as e:
    print(type(e).__name__, e)
a.setblocking(False)
try:
    a.recv(10)
except BlockingIOError as e:
    print("BlockingIOError", e.errno == errno.EAGAIN)
print(a.gettimeout(), a.getblocking())
a.close()
b.close()

# socketpair, makefile
x, y = socket.socketpair()
x.sendall(b"line one\nline two\n")
f = y.makefile("rb")
print(f.readline(), f.readline())
x.close()
print(f.read(), y.recv(5))
f.close()
y.close()
x, y = socket.socketpair()
w = x.makefile("w")
w.write("text over socket\n")
w.flush()
r = y.makefile("r")
print(repr(r.readline()))
x.close()
y.close()

# errors
s = socket.socket()
try:
    s.connect(("127.0.0.1", 1))
except ConnectionRefusedError as e:
    print("refused", e.errno == errno.ECONNREFUSED)
print(s.connect_ex(("127.0.0.1", 1)) == errno.ECONNREFUSED)
try:
    s.connect(("no-such-host.invalid", 80))
except socket.gaierror as e:
    print("gaierror", e.errno)
try:
    s.bind(("127.0.0.1", 70000))
except OverflowError as e:
    print(e)
s.close()
try:
    socket.create_connection(("127.0.0.1", 1), timeout=1)
except OSError as e:
    print(type(e).__name__)

# helpers
print(socket.inet_aton("1.2.3.4"), socket.inet_ntoa(b"\x01\x02\x03\x04"), socket.inet_pton(socket.AF_INET6, "::1"), socket.inet_ntop(socket.AF_INET6, b"\x00" * 15 + b"\x01"))
print(socket.htons(1), socket.ntohs(256), socket.htonl(1), socket.ntohl(16777216))
print([(int(a), int(b), p, c, sa) for a, b, p, c, sa in socket.getaddrinfo("127.0.0.1", 80, socket.AF_INET, socket.SOCK_STREAM)])
print(socket.gethostbyname("localhost"), socket.getservbyname("http"), socket.getprotobyname("tcp"), isinstance(socket.gethostname(), str))
print(socket.AF_INET, socket.SOCK_STREAM, socket.SOL_SOCKET, socket.SO_REUSEADDR, socket.IPPROTO_TCP, socket.TCP_NODELAY, socket.SHUT_RDWR, socket.MSG_PEEK)
print(socket.getdefaulttimeout())
socket.setdefaulttimeout(2.5)
s = socket.socket()
print(s.gettimeout(), socket.getdefaulttimeout())
s.close()
socket.setdefaulttimeout(None)
try:
    socket.inet_aton("not an ip")
except OSError as e:
    print(e)
srv = socket.create_server(("127.0.0.1", 0))
print(srv.getsockname()[0], srv.getsockopt(socket.SOL_SOCKET, socket.SO_ACCEPTCONN))
srv.close()
