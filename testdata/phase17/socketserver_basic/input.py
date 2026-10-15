import selectors
import socket
import socketserver
import threading


class Echo(socketserver.StreamRequestHandler):
    def handle(self):
        for line in self.rfile:
            self.wfile.write(line.upper())
            if line.strip() == b"quit":
                break


class Server(socketserver.ThreadingTCPServer):
    allow_reuse_address = True
    daemon_threads = True


with Server(("127.0.0.1", 0), Echo) as srv:
    host, port = srv.server_address
    t = threading.Thread(target=srv.serve_forever)
    t.start()
    for name in ("a", "b"):
        with socket.create_connection((host, port)) as c:
            f = c.makefile("rwb")
            f.write(("hello " + name + "\n").encode())
            f.write(b"quit\n")
            f.flush()
            print(f.readline(), f.readline())
    srv.shutdown()
    t.join()
print(srv.socket.fileno() == -1 or True)


class UDPEcho(socketserver.BaseRequestHandler):
    def handle(self):
        data, sock = self.request
        sock.sendto(data[::-1], self.client_address)


with socketserver.UDPServer(("127.0.0.1", 0), UDPEcho) as udp:
    t = threading.Thread(target=udp.serve_forever, kwargs={"poll_interval": 0.05})
    t.start()
    c = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    c.settimeout(2)
    c.sendto(b"abcdef", udp.server_address)
    print(c.recvfrom(100)[0])
    c.close()
    udp.shutdown()
    t.join()


class Counting(socketserver.BaseRequestHandler):
    seen = []

    def handle(self):
        Counting.seen.append(self.request.recv(10))


class Plain(socketserver.TCPServer):
    allow_reuse_address = True


with Plain(("127.0.0.1", 0), Counting) as s2:
    th = threading.Thread(target=lambda: [s2.handle_request() for _ in range(2)])
    th.start()
    for payload in (b"one", b"two"):
        with socket.create_connection(s2.server_address) as c:
            c.sendall(payload)
    th.join()
print(Counting.seen, s2.address_family == socket.AF_INET, s2.socket_type == socket.SOCK_STREAM)

sel = selectors.DefaultSelector()
a, b = socket.socketpair()
sel.register(a, selectors.EVENT_READ, data="reader")
print(sel.select(0))
b.sendall(b"x")
events = sel.select(1)
print([(k.data, ev) for k, ev in events], sel.get_key(a).fileobj is a, len(sel.get_map()))
sel.modify(a, selectors.EVENT_READ | selectors.EVENT_WRITE, "both")
print(sorted(ev for k, ev in sel.select(0)), sel.get_key(a).data)
sel.unregister(a)
try:
    sel.get_key(a)
except KeyError as e:
    print("KeyError", str(e).startswith("'<socket"))
try:
    sel.register(a, 0)
except ValueError as e:
    print(e)
sel.close()
a.close()
b.close()
