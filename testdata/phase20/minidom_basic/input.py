from xml.dom import minidom
import xml.dom

doc = minidom.parseString('<root a="1"><item id="x">first</item><item id="y">second<b>bold</b></item><!-- note --></root>')
root = doc.documentElement
print(root.tagName, root.getAttribute("a"), root.hasAttribute("zz"), len(root.childNodes))
for item in root.getElementsByTagName("item"):
    print(item.getAttribute("id"), item.firstChild.data, item.nodeType == xml.dom.Node.ELEMENT_NODE)
print([n.nodeType for n in root.childNodes])
new = doc.createElement("item")
new.setAttribute("id", "z")
new.appendChild(doc.createTextNode("third & last"))
root.appendChild(new)
print(root.toxml())
print(doc.toprettyxml(indent="  "))
d2 = minidom.Document()
top = d2.createElement("top")
d2.appendChild(top)
top.appendChild(d2.createComment("c"))
print(d2.toxml())
