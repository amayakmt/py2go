import imaplib
import socket
import threading
import time

print(imaplib.Time2Internaldate(time.struct_time((2024, 3, 4, 5, 6, 7, 0, 64, 0))) [:20])
print(imaplib.Internaldate2tuple(b'25 (INTERNALDATE "01-Jan-2024 10:20:30 +0000")')[:6])
print(imaplib.ParseFlags(b'1 (FLAGS (\\Seen \\Answered))'))
print(imaplib.Int2AP(1234))

listener = socket.socket()
listener.bind(("127.0.0.1", 0))
listener.listen()
port = listener.getsockname()[1]
seen = []


def serve():
    conn, _ = listener.accept()
    f = conn.makefile("rwb")

    def send(s):
        f.write(s.encode() + b"\r\n")
        f.flush()

    send("* OK [CAPABILITY IMAP4rev1 AUTH=PLAIN] fake ready")
    while True:
        line = f.readline()
        if not line:
            break
        tag, cmd, *rest = line.decode().rstrip("\r\n").split(" ", 2)
        seen.append(cmd)
        if cmd == "CAPABILITY":
            send("* CAPABILITY IMAP4rev1 AUTH=PLAIN")
            send(tag + " OK done")
        elif cmd == "LOGIN":
            send(tag + " OK logged in")
        elif cmd == "SELECT":
            send("* 2 EXISTS")
            send("* FLAGS (\\Seen)")
            send(tag + " OK [READ-WRITE] selected")
        elif cmd == "SEARCH":
            send("* SEARCH 1 2")
            send(tag + " OK search done")
        elif cmd == "FETCH":
            body = "Subject: hi\r\n\r\nhello"
            send("* 1 FETCH (RFC822 {%d}" % len(body))
            f.write(body.encode())
            send(")")
            send(tag + " OK fetched")
        elif cmd == "LIST":
            send('* LIST (\\HasNoChildren) "/" INBOX')
            send(tag + " OK listed")
        elif cmd == "LOGOUT":
            send("* BYE")
            send(tag + " OK bye")
            break
        else:
            send(tag + " BAD unknown")
    f.close()
    conn.close()


t = threading.Thread(target=serve)
t.start()
m = imaplib.IMAP4("127.0.0.1", port)
print(m.welcome, sorted(m.capabilities))
print(m.login("me", "pw"))
print(m.select("INBOX"))
print(m.search(None, "ALL"))
typ, data = m.fetch("1", "(RFC822)")
print(typ, data)
print(m.list())
try:
    m.noop()
except imaplib.IMAP4.error as e:
    print("error", e)
print(m.logout())
t.join()
listener.close()
print(seen)
print(issubclass(imaplib.IMAP4_SSL, imaplib.IMAP4))
