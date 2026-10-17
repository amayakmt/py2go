import email.message
import mailbox
import os
import tempfile

d = tempfile.mkdtemp()

mb = mailbox.mbox(os.path.join(d, "box.mbox"))
for i in range(3):
    m = email.message.EmailMessage()
    m["From"] = f"user{i}@example.com"
    m["Subject"] = f"hello {i}"
    m.set_content(f"body {i}\nFrom the start of a line\n")
    mb.add(m)
mb.flush()
print(len(mb), sorted(msg["Subject"] for msg in mb))
key = list(mb.keys())[1]
mb.remove(key)
mb.flush()
print(len(mb), [msg["Subject"] for msg in mb])
first = mb[list(mb.keys())[0]]
print(type(first).__name__, first.get_from().split()[0], first.get_payload().splitlines())
mb.close()
with open(os.path.join(d, "box.mbox")) as f:
    print(f.read().count(">From the start"))

md = mailbox.Maildir(os.path.join(d, "maildir"))
k = md.add("Subject: in maildir\n\ntext\n")
msg = md[k]
print(type(msg).__name__, msg["Subject"], msg.get_subdir(), msg.get_flags())
msg.set_flags("RS")
md[k] = msg
print(md[k].get_flags(), md[k].get_subdir(), len(md))
sub = md.add_folder("work")
sub.add("Subject: in folder\n\nx\n")
print(md.list_folders(), len(md.get_folder("work")))
md.lock()
md.unlock()

mh = mailbox.MH(os.path.join(d, "mh"))
k1 = mh.add("Subject: mh one\n\n1\n")
k2 = mh.add("Subject: mh two\n\n2\n")
print(sorted(mh.keys()), mh[k2]["Subject"], mh.get_sequences())
mh.set_sequences({"unseen": [k1]})
print(mh.get_sequences())

mm = mailbox.mboxMessage("From: a@b\nSubject: s\nStatus: RO\n\nbody\n")
print(mm.get_flags())
mdm = mailbox.MaildirMessage(mm)
print(mdm.get_flags(), mdm.get_subdir())
