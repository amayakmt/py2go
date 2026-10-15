import urllib.robotparser

rp = urllib.robotparser.RobotFileParser()
rp.parse("""User-agent: *
Disallow: /private/
Allow: /private/ok.html
Crawl-delay: 3
Request-rate: 2/10

User-agent: badbot
Disallow: /

Sitemap: http://example.com/sitemap.xml
""".splitlines())
for ua, url in [("anybot", "/"), ("anybot", "/private/x"), ("anybot", "/private/ok.html"), ("badbot", "/"), ("anybot", "http://example.com/a?b=1")]:
    print(ua, url, rp.can_fetch(ua, url))
print(rp.crawl_delay("anybot"), rp.request_rate("anybot"), rp.crawl_delay("badbot"))
print(rp.site_maps())
print(rp)
