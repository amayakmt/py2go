import xml.etree.ElementTree as ET

root = ET.fromstring('<catalog><book id="1" lang="en"><title>Go</title><price>10.5</price></book>'
                     '<book id="2"><title>Python</title><price>12</price></book></catalog>')
print(root.tag, len(root), root[0].tag, root[0].attrib, root[0].get("id"), root[0].get("missing", "-"))
for book in root.findall("book"):
    print(book.get("id"), book.find("title").text, float(book.findtext("price")))
print([e.tag for e in root.iter()])
print(root.find("./book[@id='2']/title").text, [t.text for t in root.iterfind(".//title")])

new = ET.SubElement(root, "book", id="3")
ET.SubElement(new, "title").text = "Rust"
new.set("lang", "de")
print(ET.tostring(new, encoding="unicode"))
root.remove(root[0])
print(len(root), [b.get("id") for b in root])

doc = ET.Element("doc", {"version": "2"})
doc.text = "intro "
item = ET.SubElement(doc, "item")
item.text = "a < b & c"
item.tail = " after"
print(ET.tostring(doc, encoding="unicode"))
print(ET.tostring(doc))
tree = ET.ElementTree(doc)
print(tree.getroot().tag)

ns = ET.fromstring('<r xmlns="urn:d" xmlns:x="urn:x"><x:a k="v">t</x:a><b/></r>')
print(ns.tag, [c.tag for c in ns], ns[0].attrib)
print(ns.find("{urn:x}a").text, ns.find("x:a", {"x": "urn:x"}).text)
try:
    ET.fromstring("<a><b></a>")
except ET.ParseError as e:
    print("ParseError", e, e.position)
el = ET.XML("<p>Hello <b>bold</b> world</p>")
print("".join(el.itertext()))
ET.indent(doc)
print(ET.tostring(doc, encoding="unicode"))
