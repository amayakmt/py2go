import http.cookiejar as cj
import urllib.request

print(cj.http2time("Wed, 09 Feb 1994 22:23:32 GMT"))
print(cj.time2isoz(0), cj.time2netscape(86400))
print(cj.split_header_words(['foo="bar"; port="80,81"; discard, bar=baz']))
print(cj.join_header_words([[("foo", "bar"), ("secure", None)], [("x", "1")]]))
print(cj.domain_match("www.example.com", ".example.com"), cj.domain_match("a.org", "b.org"))
print(cj.eff_request_host(urllib.request.Request("http://example.com/x")))

jar = cj.CookieJar()
c = cj.Cookie(0, "sid", "abc", None, False, "example.com", True, False, "/", True, False, None, True, None, None, {})
jar.set_cookie(c)
print(len(jar), [(x.name, x.value, x.domain, x.path) for x in jar])
req = urllib.request.Request("http://example.com/page")
jar.add_cookie_header(req)
print(req.get_header("Cookie"))
jar.clear()
print(len(jar))

import email.message


class FakeResponse:
    def __init__(self, headers):
        self._h = headers

    def info(self):
        return self._h


m = email.message.Message()
m["Set-Cookie"] = "a=1; Path=/; Max-Age=3600"
m["Set-Cookie"] = "b=2; Path=/x; HttpOnly"
jar2 = cj.CookieJar()
req = urllib.request.Request("http://example.com/x/y")
jar2.extract_cookies(FakeResponse(m), req)
print(len(jar2), sorted((c.name, c.value, c.path, c.domain, c.expires is None) for c in jar2))
req2 = urllib.request.Request("http://example.com/x/z")
jar2.add_cookie_header(req2)
print(req2.get_header("Cookie"))
req3 = urllib.request.Request("http://example.com/other")
jar2.add_cookie_header(req3)
print(req3.get_header("Cookie"))
jar2.clear_session_cookies()
print(len(jar2))
