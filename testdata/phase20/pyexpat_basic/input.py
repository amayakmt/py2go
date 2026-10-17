import pyexpat
from xml.parsers import expat

events = []
p = expat.ParserCreate()
p.StartElementHandler = lambda name, attrs: events.append(("start", name, sorted(attrs.items())))
p.EndElementHandler = lambda name: events.append(("end", name))
p.CharacterDataHandler = lambda data: events.append(("data", data))
p.CommentHandler = lambda text: events.append(("comment", text))
p.ProcessingInstructionHandler = lambda t, d: events.append(("pi", t, d))
p.XmlDeclHandler = lambda v, e, s: events.append(("decl", v, e, s))
p.buffer_text = True
p.Parse(b'<?xml version="1.0" encoding="utf-8"?>\n<!-- top --><root a="1" b=\'two\'>'
        b'text &amp; &lt;more&gt; &#65;&#x42;<child/><?go now?><![CDATA[<raw>]]></root>', True)
for e in events:
    print(e)

ns_events = []
q = expat.ParserCreate(None, "}")
q.StartElementHandler = lambda n, a: ns_events.append(("start", n, a))
q.EndElementHandler = lambda n: ns_events.append(("end", n))
q.StartNamespaceDeclHandler = lambda pre, uri: ns_events.append(("ns", pre, uri))
q.EndNamespaceDeclHandler = lambda pre: ns_events.append(("endns", pre))
q.ordered_attributes = True
q.Parse('<x:a xmlns:x="urn:x" xmlns="urn:d" x:k="v" plain="p"><b/></x:a>', True)
for e in ns_events:
    print(e)

for doc in [b'<a><b></a>', b'<a>', b'<a></a><b/>', b'<a>&foo;</a>', b'junk<a/>', b'<a x="1" x="2"/>', b'<a', b'', b'<a>\n  <b>\n</a>']:
    r = expat.ParserCreate()
    try:
        r.Parse(doc, True)
        print("ok", doc)
    except expat.ExpatError as e:
        print(repr(doc), e.code, e, e.lineno, e.offset, expat.ErrorString(e.code), expat.errors.messages[e.code])

d = expat.ParserCreate()
d.StartDoctypeDeclHandler = lambda n, s, p_, h: print("doctype", n, s, p_, h)
d.CharacterDataHandler = lambda t: print("data", repr(t))
d.Parse('<!DOCTYPE r [<!ENTITY who "world">]><r>hello &who;</r>', True)
print(pyexpat.EXPAT_VERSION, expat.errors.XML_ERROR_SYNTAX, expat.errors.codes[expat.errors.XML_ERROR_SYNTAX])
