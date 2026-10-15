from urllib import parse
from urllib.parse import urlparse, urlsplit, urlunparse, urljoin, urlencode, quote, unquote, quote_plus, unquote_plus, parse_qs, parse_qsl, urldefrag

u = urlparse("https://user:pw@example.com:8080/a/b;p?x=1&y=2#frag")
print(u, u.scheme, u.netloc, u.path, u.params, u.query, u.fragment, u.username, u.password, u.hostname, u.port)
print(urlsplit("//host/path?q"), urlsplit("mailto:a@b.c"), urlsplit("/just/path"), urlparse(""))
print(urlunparse(u), u._replace(path="/z").geturl(), urlparse("HTTP://EXAMPLE.com/").scheme)
print(urljoin("http://a/b/c/d;p?q", "../g"), urljoin("http://a/b/c/d", "//other/x"), urljoin("http://a/b/", "?y"), urljoin("http://a/b", "#f"), urljoin("http://a", "x"))
print(urlencode({"a": 1, "b": "x y", "c": ["1", "2"]}, doseq=True), urlencode([("k", "é"), ("k2", "a&b")]), urlencode({"q": "a b"}, quote_via=quote))
print(quote("/a b/ü?"), quote("/a b/", safe=""), quote_plus("a b&c"), unquote("%41%20%C3%BC%zz"), unquote_plus("a+b%20c"), quote(b"\xff\x00"))
print(parse_qs("a=1&a=2&b=&c"), parse_qs("a=1&b=", keep_blank_values=True), parse_qsl("x=%20y&z=1;w=2"), parse_qs("a=1;b=2", separator=";"))
print(urldefrag("http://a/b#c"), parse.splitquery("a?b") if False else "")
for bad in ["http://[::1/", "http://a:99999999999/"]:
    try:
        r = urlparse(bad)
        print(r.port)
    except ValueError as e:
        print(e)
print(urlparse("http://[::1]:80/").hostname, urlparse("http://h:80").port, urlparse("x://h:abc").netloc)
print(parse.urlsplit("http://a.com/ü").path, parse.quote("ü"), parse.unquote("%C3%BC"), parse.unquote_to_bytes("%C3%BC"))
