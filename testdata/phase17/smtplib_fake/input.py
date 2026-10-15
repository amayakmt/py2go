import smtplib
import socket
import threading
from email.message import EmailMessage

received = []


def serve(ls):
    conn, _ = ls.accept()
    f = conn.makefile("rwb", buffering=0)
    f.write(b"220 fake.example ESMTP ready\r\n")
    mail = {}
    while True:
        line = f.readline()
        if not line:
            break
        text = line.decode().rstrip("\r\n")
        up = text.upper()
        if up.startswith("EHLO"):
            f.write(b"250-fake.example\r\n250-SIZE 100000\r\n250 8BITMIME\r\n")
        elif up.startswith("MAIL FROM"):
            mail = {"from": text[10:], "to": []}
            f.write(b"250 OK\r\n")
        elif up.startswith("RCPT TO"):
            if "bad" in text:
                f.write(b"550 no such user\r\n")
            else:
                mail["to"].append(text[8:])
                f.write(b"250 OK\r\n")
        elif up == "DATA":
            f.write(b"354 go ahead\r\n")
            lines = []
            while True:
                l = f.readline().decode().rstrip("\r\n")
                if l == ".":
                    break
                lines.append(l[1:] if l.startswith("..") else l)
            mail["data"] = lines
            received.append(mail)
            f.write(b"250 queued\r\n")
        elif up == "NOOP":
            f.write(b"250 OK\r\n")
        elif up == "RSET":
            f.write(b"250 OK\r\n")
        elif up == "QUIT":
            f.write(b"221 bye\r\n")
            break
        else:
            f.write(b"500 what\r\n")
    f.close()
    conn.close()


ls = socket.socket()
ls.bind(("127.0.0.1", 0))
ls.listen(1)
port = ls.getsockname()[1]
t = threading.Thread(target=serve, args=(ls,))
t.start()

s = smtplib.SMTP("127.0.0.1", port, timeout=5, local_hostname="client.example")
print(s.ehlo())
print(s.has_extn("size"), s.has_extn("starttls"), sorted(s.esmtp_features))
print(s.noop())
try:
    s.sendmail("a@x.org", ["good@x.org", "bad@x.org"], "Subject: hi\r\n\r\n.dot line\r\nbody\r\n")
    print("sent")
except smtplib.SMTPRecipientsRefused as e:
    print("refused", e.recipients)
msg = EmailMessage()
msg["From"] = "me@x.org"
msg["To"] = "you@x.org"
msg["Subject"] = "hello"
msg.set_content("body text\n")
print(s.send_message(msg))
print(s.quit())
t.join()
ls.close()
for m in received:
    print(m["from"], m["to"])
    for l in m["data"]:
        print("  |", l)
print(smtplib.quoteaddr("Me <me@x.org>"), smtplib.quotedata("a\n.b\r\nc"))
