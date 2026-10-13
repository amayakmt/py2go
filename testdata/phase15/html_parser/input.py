from html.parser import HTMLParser


class Logger(HTMLParser):
    def __init__(self, **kw):
        HTMLParser.__init__(self, **kw)
        self.events = []

    def handle_starttag(self, tag, attrs):
        self.events.append(("start", tag, attrs, self.getpos()))

    def handle_startendtag(self, tag, attrs):
        self.events.append(("startend", tag, attrs))

    def handle_endtag(self, tag):
        self.events.append(("end", tag))

    def handle_data(self, data):
        self.events.append(("data", data))

    def handle_comment(self, data):
        self.events.append(("comment", data))

    def handle_decl(self, decl):
        self.events.append(("decl", decl))

    def handle_pi(self, data):
        self.events.append(("pi", data))

    def handle_entityref(self, name):
        self.events.append(("entityref", name))

    def handle_charref(self, name):
        self.events.append(("charref", name))

    def unknown_decl(self, data):
        self.events.append(("unknown_decl", data))


def run(text, chunks=None, **kw):
    p = Logger(**kw)
    if chunks:
        for i in range(0, len(text), chunks):
            p.feed(text[i:i + chunks])
    else:
        p.feed(text)
    p.close()
    return p.events


docs = [
    '<!DOCTYPE html><html lang="en"><head><title>T &amp; U</title></head><body class="a b" id=main data-x=\'1\' hidden><p>Hello &lt;world&gt; &#65; &#x42; &copy; &nosuch;</p><br/><img src="a.png" alt=""/><!-- a comment --><?php echo 1 ?></body></html>',
    "<script>if (a < b && c > d) { x = '</scr' + 'ipt>'; }</script><style>p > a { color: red }</style><textarea>&amp; <b>x</b></textarea><title>a &amp; b</title>",
    "<a href=x onclick = \"f('y')\" disabled>link</a><B>bold</B><P>unclosed<p>second",
    "text only &amp; entity &#169; and & bare < lone",
    "<div\n  class='x'\n  >multi\nline</div  >",
    "<![CDATA[ not really ]]><![if IE]>cond<![endif]><!ELEMENT x ANY><!-->empty<!--->also<!-- a -- b -->done",
    "<a b=1 b=2 c d='' e=\"\"=>x</a><input value=\"&amp;&lt;\" title='&quot;'>",
    "</p></ BAD><//><x/ y>text<z / >",
    "<?xml version='1.0'?><root/>",
]
for d in docs:
    print(run(d))
    print(run(d, chunks=7))
print(run(docs[0], convert_charrefs=False))
print(run("<noscript><b>x</b></noscript>", scripting=True), run("<noscript><b>x</b></noscript>"))
p = Logger()
p.feed("<p id=1>one</p>")
print(p.get_starttag_text(), p.getpos(), p.rawdata, p.lasttag, p.cdata_elem)
p.reset()
p.feed("<x>")
p.close()
print(p.events, p.get_starttag_text())
print(HTMLParser.CDATA_CONTENT_ELEMENTS, HTMLParser().convert_charrefs)
