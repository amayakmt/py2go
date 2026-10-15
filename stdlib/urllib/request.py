"""An extensible library for opening URLs (reduced hand-written port).

Supports http:, file: and data: URLs, redirects, and basic opener/handler
plumbing.  No ssl, proxies, ftp or authentication handlers.
"""
import base64
import http.client
import io
import os
import socket
import sys
from urllib.error import URLError, HTTPError, ContentTooShortError
from urllib.parse import urlparse, urlsplit, urljoin, unquote, quote, urlunparse, unquote_to_bytes, splittype as _splittype, splithost as _splithost
from urllib.response import addinfourl, addclosehook

__all__ = ['Request', 'OpenerDirector', 'BaseHandler', 'HTTPRedirectHandler', 'HTTPHandler', 'HTTPCookieProcessor', 'FileHandler', 'DataHandler', 'build_opener', 'install_opener', 'urlopen', 'urlretrieve', 'pathname2url', 'url2pathname', 'quote', 'unquote', 'URLError', 'HTTPError', 'getproxies']

MAXREDIRECTS = 10


def _obj(x: object) -> object:
    return x


def pathname2url(pathname):
    return quote(pathname)


def url2pathname(pathname):
    return unquote(pathname)


def getproxies():
    return {}


class Request:

    def __init__(self, url, data=None, headers=None, origin_req_host=None, unverifiable=False, method=None):
        self.full_url = url
        self.headers = {}
        self.unredirected_hdrs = {}
        self._data = _obj(None)
        self.data = _obj(data)
        self._tunnel_host = None
        if headers is None:
            headers = {}
        for key in headers:
            self.add_header(key, headers[key])
        if origin_req_host is None:
            origin_req_host = urlparse(url).hostname or ''
        self.origin_req_host = origin_req_host
        self.unverifiable = unverifiable
        self._method = _obj(method)
        self.timeout = _obj(None)

    @property
    def full_url(self):
        if self.fragment:
            return '{}#{}'.format(self._full_url, self.fragment)
        return self._full_url

    @full_url.setter
    def full_url(self, url):
        self._full_url = _obj(url)
        pos = url.find('#')
        if pos >= 0:
            self.fragment = _obj(url[pos + 1:])
            self._full_url = _obj(url[:pos])
        else:
            self.fragment = _obj(None)
        self._parse()

    def _parse(self):
        self.type = _obj(None)
        t, rest = _splittype(self._full_url)
        if t is None:
            raise ValueError('unknown url type: %r' % self.full_url)
        self.type = _obj(t.lower())
        host, selector = _splithost(rest)
        self.host = _obj(unquote(host) if host else host)
        self.selector = _obj(selector)

    @property
    def data(self):
        return self._data

    @data.setter
    def data(self, data):
        self._data = _obj(data)
        if data is not None:
            if self.has_header('Content-length'):
                self.remove_header('Content-length')

    def get_method(self):
        if self._method is not None:
            return self._method
        if self.data is not None:
            return 'POST'
        return 'GET'

    def get_full_url(self):
        return self.full_url

    def has_proxy(self):
        return False

    def add_header(self, key, val):
        self.headers[key.capitalize()] = val

    def add_unredirected_header(self, key, val):
        self.unredirected_hdrs[key.capitalize()] = val

    def has_header(self, header_name):
        return header_name in self.headers or header_name in self.unredirected_hdrs

    def get_header(self, header_name, default=None):
        return self.headers.get(header_name, self.unredirected_hdrs.get(header_name, default))

    def remove_header(self, header_name):
        if header_name in self.headers:
            del self.headers[header_name]
        if header_name in self.unredirected_hdrs:
            del self.unredirected_hdrs[header_name]

    def header_items(self):
        hdrs = {}
        for k in self.unredirected_hdrs:
            hdrs[k] = self.unredirected_hdrs[k]
        for k in self.headers:
            hdrs[k] = self.headers[k]
        return list(hdrs.items())

    @property
    def host_name(self):
        return self.host


class BaseHandler:
    handler_order = 500

    def add_parent(self, parent):
        self.parent = _obj(parent)

    def close(self):
        pass

    def __lt__(self, other):
        return self.handler_order < other.handler_order


class HTTPRedirectHandler(BaseHandler):
    max_repeats = 4
    max_redirections = 10
    handler_order = 600

    def redirect_request(self, req, fp, code, msg, headers, newurl):
        m = req.get_method()
        if not (code in (301, 302, 303, 307, 308) and m in ('GET', 'HEAD') or (code in (301, 302, 303) and m == 'POST')):
            raise HTTPError(req.full_url, code, msg, headers, fp)
        newurl = newurl.replace(' ', '%20')
        newheaders = {}
        for k, v in req.headers.items():
            if k.lower() not in ('content-length', 'content-type'):
                newheaders[k] = v
        data = req.data
        method = req._method
        if code in (301, 302, 303) and m == 'POST':
            data = None
            method = None
        elif code in (307, 308):
            pass
        return Request(newurl, data=data, headers=newheaders, origin_req_host=req.origin_req_host, unverifiable=True, method=method)

    def http_error_redirect(self, req, fp, code, msg, headers):
        if 'location' in headers:
            newurl = headers['location']
        elif 'uri' in headers:
            newurl = headers['uri']
        else:
            return None
        urlparts = urlparse(newurl)
        if urlparts.scheme not in ('http', 'https', 'ftp', ''):
            raise HTTPError(newurl, code, "%s - Redirection to url '%s' is not allowed" % (msg, newurl), headers, fp)
        if not urlparts.path and urlparts.netloc:
            urlparts = urlparts._replace(path='/')
            newurl = urlunparse(urlparts)
        newurl = urljoin(req.full_url, newurl)
        new = self.redirect_request(req, fp, code, msg, headers, newurl)
        if new is None:
            return None
        return self.parent.open(new, timeout=req.timeout)


class HTTPHandler(BaseHandler):

    def http_open(self, req):
        host = req.host
        if not host:
            raise URLError('no host given')
        h = http.client.HTTPConnection(host, timeout=req.timeout if req.timeout is not None else socket._GLOBAL_DEFAULT_TIMEOUT)
        headers = {}
        for k, v in req.header_items():
            headers[k] = v
        if not req.has_header('Host'):
            headers['Host'] = host
        headers['Connection'] = 'close'
        headers = {name.title(): val for name, val in headers.items()}
        try:
            try:
                h.request(req.get_method(), req.selector, req.data, headers)
                r = h.getresponse()
            except OSError as err:
                raise URLError(err)
        except URLError:
            h.close()
            raise
        resp = addinfourl(r, r.msg, req.get_full_url(), r.status)
        resp.reason = r.reason
        resp.msg = r.reason
        return resp


class HTTPCookieProcessor(BaseHandler):

    def __init__(self, cookiejar=None):
        import http.cookiejar
        if cookiejar is None:
            cookiejar = http.cookiejar.CookieJar()
        self.cookiejar = _obj(cookiejar)


class DataHandler(BaseHandler):

    def data_open(self, req):
        url = req.full_url
        scheme, data = url.split(':', 1)
        mediatype, data = data.split(',', 1)
        data = unquote_to_bytes(data)
        if mediatype.endswith(';base64'):
            data = base64.decodebytes(data)
            mediatype = mediatype[:-7]
        if not mediatype:
            mediatype = 'text/plain;charset=US-ASCII'
        import email.message
        headers = email.message.Message()
        headers['Content-type'] = mediatype
        headers['Content-length'] = str(len(data))
        return addinfourl(io.BytesIO(data), headers, url)


class FileHandler(BaseHandler):

    def file_open(self, req):
        url = req.selector
        if url[:2] == '//' and url[2:3] != '/' and (req.host and req.host != 'localhost'):
            raise URLError('file:// scheme is supported only on localhost')
        return self.open_local_file(req)

    def open_local_file(self, req):
        import email.message
        import mimetypes
        host = req.host
        filename = req.selector
        localfile = url2pathname(filename)
        try:
            stats = os.stat(localfile)
            size = stats.st_size
            mtype = mimetypes.guess_type(filename)[0]
            headers = email.message.Message()
            headers['Content-type'] = mtype or 'text/plain'
            headers['Content-length'] = '%d' % size
            if host:
                if host != 'localhost':
                    raise URLError('file not on local host')
            return addinfourl(open(localfile, 'rb'), headers, 'file://' + host + filename if host else 'file://' + filename)
        except OSError as exp:
            raise URLError(exp)


class OpenerDirector:

    def __init__(self):
        self.handlers = []
        self.addheaders = [('User-agent', 'Python-urllib/3.13')]

    def add_handler(self, handler):
        handler.add_parent(self)
        self.handlers.append(handler)
        self.handlers.sort()

    def close(self):
        pass

    def open(self, fullurl, data=None, timeout=None):
        if isinstance(fullurl, str):
            req = Request(fullurl, data)
        else:
            req = fullurl
            if data is not None:
                req.data = data
        req.timeout = timeout
        protocol = req.type
        if protocol == 'http':
            response = self._open_http(req)
        elif protocol == 'file':
            response = self._open_file(req)
        elif protocol == 'data':
            response = self._open_data(req)
        else:
            raise URLError('unknown url type: %s' % protocol)
        return response

    def _find(self, kind):
        for h in self.handlers:
            if kind == 'http' and isinstance(h, HTTPHandler):
                return h
            if kind == 'file' and isinstance(h, FileHandler):
                return h
            if kind == 'data' and isinstance(h, DataHandler):
                return h
        return None

    def _open_http(self, req):
        h = self._find('http')
        if h is None:
            raise URLError('unknown url type: http')
        for k, v in self.addheaders:
            if not req.has_header(k):
                req.add_unredirected_header(k, v)
        if req.data is not None and not req.has_header('Content-type'):
            req.add_unredirected_header('Content-type', 'application/x-www-form-urlencoded')
        if req.data is not None and not req.has_header('Content-length'):
            req.add_unredirected_header('Content-length', '%d' % len(req.data))
        jar = None
        for hh in self.handlers:
            if isinstance(hh, HTTPCookieProcessor):
                jar = hh.cookiejar
        if jar is not None:
            jar.add_cookie_header(req)
        response = h.http_open(req)
        if jar is not None:
            jar.extract_cookies(response, req)
        code = response.code
        if not (200 <= code < 300):
            hdrs = response.info()
            if code in (301, 302, 303, 307, 308):
                redirector = None
                for hh in self.handlers:
                    if isinstance(hh, HTTPRedirectHandler):
                        redirector = hh
                if redirector is not None:
                    result = redirector.http_error_redirect(req, response, code, response.msg, hdrs)
                    if result is not None:
                        return result
            raise HTTPError(req.full_url, code, response.msg, hdrs, response)
        return response

    def _open_file(self, req):
        h = self._find('file')
        if h is None:
            raise URLError('unknown url type: file')
        return h.file_open(req)

    def _open_data(self, req):
        h = self._find('data')
        if h is None:
            raise URLError('unknown url type: data')
        return h.data_open(req)


def build_opener(*handlers):
    opener = OpenerDirector()
    default_handlers = [HTTPHandler(), HTTPRedirectHandler(), FileHandler(), DataHandler()]
    for h in default_handlers:
        skip = False
        for check in handlers:
            if isinstance(check, h.__class__):
                skip = True
        if not skip:
            opener.add_handler(h)
    for h in handlers:
        opener.add_handler(h)
    return opener


_opener = None


def urlopen(url, data=None, timeout=socket._GLOBAL_DEFAULT_TIMEOUT, *, context=None):
    global _opener
    if _opener is None:
        _opener = build_opener()
    t = None
    if timeout is not socket._GLOBAL_DEFAULT_TIMEOUT:
        t = timeout
    return _opener.open(url, data, t)


def install_opener(opener):
    global _opener
    _opener = opener


def urlretrieve(url, filename=None, reporthook=None, data=None):
    with urlopen(url, data) as fp:
        headers = fp.info()
        if filename is None:
            raise ValueError('filename required')
        with open(filename, 'wb') as tfp:
            while True:
                block = fp.read(8192)
                if not block:
                    break
                tfp.write(block)
    return filename, headers
