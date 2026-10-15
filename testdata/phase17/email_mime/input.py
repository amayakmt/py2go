import email
import email.utils
from email import encoders, policy
from email.header import Header, decode_header, make_header
from email.message import EmailMessage
from email.mime.application import MIMEApplication
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from email.parser import BytesParser, Parser
from email.generator import Generator, BytesGenerator
import io

msg = MIMEMultipart("mixed", boundary="BOUNDARY123")
msg["Subject"] = "Report"
msg["From"] = "me@example.com"
msg["To"] = "you@example.com"
msg["Date"] = "Tue, 02 Jan 2024 03:04:05 +0000"
msg.attach(MIMEText("plain body\n", "plain"))
msg.attach(MIMEText("<p>html</p>", "html", "utf-8"))
part = MIMEApplication(b"\x00\x01binary data", Name="data.bin")
part.add_header("Content-Disposition", "attachment", filename="data.bin")
msg.attach(part)
text = msg.as_string()
print(text)
back = email.message_from_string(text)
print(back.is_multipart(), [p.get_content_type() for p in back.walk()], back.get_boundary())
for p in back.get_payload():
    print(p.get_content_type(), p.get_filename(), p.get_content_charset(), repr(p.get_payload(decode=True))[:40])

m = MIMEText("héllo wörld", "plain", "utf-8")
print(m.as_string())
print(m.get_payload(decode=True).decode("utf-8"), m["Content-Transfer-Encoding"])
print(MIMEText("ascii only").as_string())

h = Header("Ünïcode subject", "utf-8")
print(h.encode(), str(h), decode_header(h.encode()), str(make_header(decode_header("=?utf-8?q?caf=C3=A9?= au lait"))))
print(email.utils.formataddr(("Ünï C", "a@b.c")), email.utils.parseaddr('"Doe, John" <john@example.com>'), email.utils.getaddresses(["a@b.c, D <d@e.f>"]))
print(email.utils.parsedate("Mon, 20 Nov 1995 19:12:08 -0500"), email.utils.parsedate_tz("Mon, 20 Nov 1995 19:12:08 -0500")[9], email.utils.formatdate(0, usegmt=True), email.utils.format_datetime(email.utils.parsedate_to_datetime("Mon, 20 Nov 1995 19:12:08 -0500")))
print(email.utils.quote('a"b\\c'), email.utils.unquote('"abc"'), email.utils.decode_rfc2231("utf-8''caf%C3%A9"), email.utils.encode_rfc2231("café", "utf-8"))

em = EmailMessage()
em["From"] = "a@example.com"
em["To"] = "b@example.com"
em["Subject"] = "Modern"
em.set_content("simple text\n")
em.add_alternative("<b>rich</b>\n", subtype="html")
em.add_attachment(b"PDFDATA", maintype="application", subtype="pdf", filename="x.pdf")
print(em.get_content_type(), [p.get_content_type() for p in em.walk()], em["MIME-Version"])
print([ (p.get_content_type(), p.get_filename()) for p in em.iter_attachments()], em.get_body(preferencelist=("plain",)).get_content())
print(em.is_multipart(), len(list(em.iter_parts())))

p = Parser().parsestr("Subject: s\nContent-Type: text/plain; charset=us-ascii\n\nbody text\n")
print(p["subject"], p.get_payload(), p.get_params(), p.get_param("charset"))
bp = BytesParser().parsebytes(b"From: x@y.z\r\nContent-Type: multipart/mixed; boundary=bb\r\n\r\n--bb\r\nContent-Type: text/plain\r\n\r\npart one\r\n--bb--\r\n")
print(bp.get_content_type(), len(bp.get_payload()), bp.get_payload(0).get_payload())
out = io.StringIO()
Generator(out, mangle_from_=False, maxheaderlen=40).flatten(msg.get_payload(0))
print(out.getvalue())
bio = io.BytesIO()
BytesGenerator(bio).flatten(back)
print(len(bio.getvalue()) > 100)
p2 = MIMEText("body", "plain")
encoders.encode_base64(p2)
print(p2["Content-Transfer-Encoding"], p2.get_payload())
print(email.message_from_string("Subject: x\n\nbody", policy=policy.default)["Subject"], policy.SMTP.linesep == "\r\n", policy.default.max_line_length)
