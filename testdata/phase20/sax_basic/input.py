import io
import xml.sax
from xml.sax.handler import ContentHandler
from xml.sax.saxutils import escape, unescape, quoteattr, XMLGenerator


class Collector(ContentHandler):
    def __init__(self):
        super().__init__()
        self.events = []

    def startDocument(self):
        self.events.append("startDocument")

    def endDocument(self):
        self.events.append("endDocument")

    def startElement(self, name, attrs):
        self.events.append(("start", name, sorted(attrs.items())))

    def endElement(self, name):
        self.events.append(("end", name))

    def characters(self, content):
        if content.strip():
            self.events.append(("chars", content))


h = Collector()
xml.sax.parseString(b'<doc a="1"><p>hello</p><p b="2">world &amp; more</p></doc>', h)
for e in h.events:
    print(e)
try:
    xml.sax.parseString(b"<a><b></a>", ContentHandler())
except xml.sax.SAXParseException as e:
    print("SAXParseException", e.getMessage(), e.getLineNumber(), e.getColumnNumber())
print(escape("a < b & c > d"), unescape("&lt;x&gt; &amp;"), quoteattr('say "hi"'))
out = io.StringIO()
g = XMLGenerator(out, encoding="unicode")
g.startDocument()
g.startElement("root", {"k": "v"})
g.characters("text & <stuff>")
g.endElement("root")
g.endDocument()
print(out.getvalue())
