import email
from email import message_from_string, message_from_bytes
from email.message import Message, EmailMessage

raw = """From: Alice <alice@example.com>
To: bob@example.com, "Carol, C" <carol@example.com>
Subject: Hello there
Date: Mon, 01 Jan 2024 10:00:00 +0000
Message-ID: <abc123@example.com>
Content-Type: text/plain; charset="utf-8"

Body line one
Body line two
"""
m = message_from_string(raw)
print(m["Subject"], m["from"], m.get("X-Missing", "dflt"), list(m.keys()), len(m))
print(m.get_payload(), m.get_content_type(), m.get_content_maintype(), m.get_content_subtype(), m.get_charset(), m.get_content_charset())
print(m.is_multipart(), m.items()[:2], "Subject" in m, m.get_all("To"))
m["X-Custom"] = "value"
m.replace_header("Subject", "Changed")
del m["Message-ID"]
print(m["X-Custom"], m["Subject"], m["Message-ID"])
print(m.as_string())
