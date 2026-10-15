import poplib
import socket
import threading

MSGS = [b"Subject: one\r\n\r\nhello\r\n", b"Subject: two\r\n\r\nworld\r\nline2\r\n"]


def serve(ls):
    conn, _ = ls.accept()
    f = conn.makefile("rwb", buffering=0)
    f.write(b"+OK fake pop3 ready\r\n")
    while True:
        line = f.readline()
        if not line:
            break
        cmd = line.decode().strip().split()
        c = cmd[0].upper()
        if c == "USER":
            f.write(b"+OK\r\n")
        elif c == "PASS":
            f.write(b"+OK logged in\r\n" if cmd[1] == "secret" else b"-ERR bad password\r\n")
        elif c == "STAT":
            f.write(("+OK %d %d\r\n" % (len(MSGS), sum(len(m) for m in MSGS))).encode())
        elif c == "LIST":
            f.write(b"+OK\r\n")
            for i, m in enumerate(MSGS):
                f.write(("%d %d\r\n" % (i + 1, len(m))).encode())
            f.write(b".\r\n")
        elif c == "RETR":
            m = MSGS[int(cmd[1]) - 1]
            f.write(b"+OK message follows\r\n" + m.replace(b"\r\n.", b"\r\n..") + b".\r\n")
        elif c == "DELE":
            f.write(b"+OK deleted\r\n")
        elif c == "NOOP":
            f.write(b"+OK\r\n")
        elif c == "QUIT":
            f.write(b"+OK bye\r\n")
            break
        else:
            f.write(b"-ERR unknown\r\n")
    f.close()
    conn.close()


ls = socket.socket()
ls.bind(("127.0.0.1", 0))
ls.listen(1)
port = ls.getsockname()[1]
t = threading.Thread(target=serve, args=(ls,))
t.start()

p = poplib.POP3("127.0.0.1", port, timeout=5)
print(p.getwelcome())
print(p.user("me"))
try:
    p.pass_("wrong")
except poplib.error_proto as e:
    print("error_proto", e)
print(p.pass_("secret"))
print(p.stat())
print(p.list())
print(p.retr(2))
print(p.dele(1))
print(p.noop())
print(p.quit())
t.join()
ls.close()
