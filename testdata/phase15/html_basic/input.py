import html
import html.entities

print(html.escape('<a href="x">Tom & "Jerry\'s"</a>'), html.escape("<>&\"'", quote=False), html.escape(""))
tests = ["&lt;b&gt;", "&amp;amp;", "&copy; 2024", "&copy2024", "&#65;&#x42;&#X43;", "&#0;&#128;&#150;&#xD800;&#x110000;&#1;&#x7f;&#xfffe;", "&notit;", "&notin;", "&nosuchentity;", "AT&T", "&", "a &lt b", "&lt", "&ampxyz", "&AMP;&GT;", "&euro;&eacute;&frac12;", "&NotEqualTilde;", "&#;", "&#x;", "&#99999999999;", "no entities", "&#38;lt;"]
for t in tests:
    print(repr(t), repr(html.unescape(t)))
print(len(html.entities.html5), len(html.entities.name2codepoint), html.entities.name2codepoint["eacute"], html.entities.codepoint2name[233], html.entities.entitydefs["amp"], html.entities.html5["amp;"], html.entities.html5["amp"], repr(html.entities.html5["NotEqualTilde;"]))
