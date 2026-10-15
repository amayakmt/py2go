"""urllib.parse: parse URLs into components and quote and unquote their parts."""
import re
import ipaddress

__all__ = ["urlparse", "urlunparse", "urljoin", "urldefrag", "urlsplit", "urlunsplit", "urlencode", "parse_qs",
           "parse_qsl", "quote", "quote_plus", "quote_from_bytes", "unquote", "unquote_plus", "unquote_to_bytes",
           "DefragResult", "ParseResult", "SplitResult", "DefragResultBytes", "ParseResultBytes", "SplitResultBytes"]

uses_relative = ["", "ftp", "http", "gopher", "nntp", "imap", "wais", "file", "https", "shttp", "mms", "prospero",
                 "rtsp", "rtsps", "rtspu", "sftp", "svn", "svn+ssh", "ws", "wss"]
uses_netloc = ["", "ftp", "http", "gopher", "nntp", "telnet", "imap", "wais", "file", "mms", "https", "shttp", "snews",
               "prospero", "rtsp", "rtsps", "rtspu", "rsync", "svn", "svn+ssh", "sftp", "nfs", "git", "git+ssh", "ws", "wss"]
uses_params = ["", "ftp", "hdl", "prospero", "http", "imap", "https", "shttp", "rtsp", "rtsps", "rtspu", "sip", "sips", "mms", "sftp", "tel"]
non_hierarchical = ["gopher", "hdl", "mailto", "news", "telnet", "wais", "imap", "snews", "sip", "sips"]
uses_query = ["", "http", "wais", "imap", "https", "shttp", "mms", "gopher", "rtsp", "rtsps", "rtspu", "sip", "sips"]
uses_fragment = ["", "ftp", "hdl", "http", "gopher", "news", "nntp", "wais", "https", "shttp", "snews", "file", "prospero"]

scheme_chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+-."

_WHATWG_C0_CONTROL_OR_SPACE = "\x00\x01\x02\x03\x04\x05\x06\x07\x08\t\n\x0b\x0c\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f "
_UNSAFE_URL_BYTES_TO_REMOVE = ["\t", "\r", "\n"]


def clear_cache():
    pass


def _obj(x: object) -> object:
    return x


class _Result:
    _fields = _obj(())

    def __init__(self, *args):
        if len(args) != len(self._fields):
            raise TypeError("%s.__new__() takes %d positional arguments but %d were given" % (
                type(self).__name__, len(self._fields) + 1, len(args) + 1))
        self._vals = tuple(args)

    def __getitem__(self, i):
        return self._vals[i]

    def __iter__(self):
        return iter(self._vals)

    def __len__(self):
        return len(self._vals)

    def __eq__(self, other):
        if isinstance(other, _Result):
            return self._vals == other._vals
        return self._vals == other

    def __lt__(self, other):
        return self._vals < tuple(other)

    def __hash__(self):
        return hash(self._vals)

    def count(self, x):
        return self._vals.count(x)

    def index(self, x):
        return self._vals.index(x)

    def _asdict(self):
        return dict(zip(self._fields, self._vals))

    def __repr__(self):
        parts = []
        for k, v in zip(self._fields, self._vals):
            parts.append("%s=%r" % (k, v))
        return "%s(%s)" % (type(self).__name__, ", ".join(parts))

    def __getnewargs__(self):
        return self._vals

    def decode(self, encoding="ascii", errors="strict"):
        return _decode_result(self, encoding, errors)

    def encode(self, encoding="ascii", errors="strict"):
        return _encode_result(self, encoding, errors)


def _userinfo_str(netloc):
    userinfo, have_info, hostinfo = netloc.rpartition("@")
    if have_info:
        username, have_password, password = userinfo.partition(":")
        if not have_password:
            password = None
    else:
        username = password = None
    return username, password


def _hostinfo_str(netloc):
    _, _, hostinfo = netloc.rpartition("@")
    _, have_open_br, bracketed = hostinfo.partition("[")
    if have_open_br:
        hostname, _, port = bracketed.partition("]")
        _, _, port = port.partition(":")
    else:
        hostname, _, port = hostinfo.partition(":")
    if not port:
        port = None
    return hostname, port


class _NetlocMixin(_Result):
    @property
    def username(self):
        n = self._netloc_text()
        r = _userinfo_str(n)[0]
        return self._wrap(r)

    @property
    def password(self):
        n = self._netloc_text()
        r = _userinfo_str(n)[1]
        return self._wrap(r)

    @property
    def hostname(self):
        hostname = _hostinfo_str(self._netloc_text())[0]
        if not hostname:
            return None
        hostname, percent, zone = hostname.partition("%")
        return self._wrap(hostname.lower() + percent + zone)

    @property
    def port(self):
        port = _hostinfo_str(self._netloc_text())[1]
        if port is not None:
            if port.isdigit() and port.isascii():
                port = int(port)
            else:
                raise ValueError(f"Port could not be cast to integer value as {self._wrap(port)!r}")
            if not (0 <= port <= 65535):
                raise ValueError("Port out of range 0-65535")
        return port

    def _netloc_text(self):
        return ""

    def _wrap(self, s):
        return s


class _StrNetloc(_NetlocMixin):
    def _netloc_text(self):
        return self._vals[1]

    def geturl(self):
        return self._geturl()

    def _geturl(self):
        return ""


class _BytesNetloc(_NetlocMixin):
    def _netloc_text(self):
        return self._vals[1].decode("ascii")

    def _wrap(self, s):
        if s is None:
            return None
        return s.encode("ascii")


class DefragResult(_Result):
    _fields = _obj(("url", "fragment"))

    @property
    def url(self):
        return self._vals[0]

    @property
    def fragment(self):
        return self._vals[1]

    def geturl(self):
        if self.fragment:
            return self.url + "#" + self.fragment
        return self.url

    def _replace(self, **kwargs):
        d = self._asdict()
        d.update(kwargs)
        return type(self)(d["url"], d["fragment"])


class DefragResultBytes(DefragResult):
    def geturl(self):
        if self.fragment:
            return self.url + b"#" + self.fragment
        return self.url


class SplitResult(_StrNetloc):
    _fields = _obj(("scheme", "netloc", "path", "query", "fragment"))

    @property
    def scheme(self):
        return self._vals[0]

    @property
    def netloc(self):
        return self._vals[1]

    @property
    def path(self):
        return self._vals[2]

    @property
    def query(self):
        return self._vals[3]

    @property
    def fragment(self):
        return self._vals[4]

    def geturl(self):
        return urlunsplit(self)

    def _replace(self, **kwargs):
        d = self._asdict()
        d.update(kwargs)
        return type(self)(d["scheme"], d["netloc"], d["path"], d["query"], d["fragment"])


class SplitResultBytes(_BytesNetloc):
    _fields = _obj(("scheme", "netloc", "path", "query", "fragment"))

    @property
    def scheme(self):
        return self._vals[0]

    @property
    def netloc(self):
        return self._vals[1]

    @property
    def path(self):
        return self._vals[2]

    @property
    def query(self):
        return self._vals[3]

    @property
    def fragment(self):
        return self._vals[4]

    def geturl(self):
        return urlunsplit(self)

    def _replace(self, **kwargs):
        d = self._asdict()
        d.update(kwargs)
        return type(self)(d["scheme"], d["netloc"], d["path"], d["query"], d["fragment"])


class ParseResult(_StrNetloc):
    _fields = _obj(("scheme", "netloc", "path", "params", "query", "fragment"))

    @property
    def scheme(self):
        return self._vals[0]

    @property
    def netloc(self):
        return self._vals[1]

    @property
    def path(self):
        return self._vals[2]

    @property
    def params(self):
        return self._vals[3]

    @property
    def query(self):
        return self._vals[4]

    @property
    def fragment(self):
        return self._vals[5]

    def geturl(self):
        return urlunparse(self)

    def _replace(self, **kwargs):
        d = self._asdict()
        d.update(kwargs)
        return type(self)(d["scheme"], d["netloc"], d["path"], d["params"], d["query"], d["fragment"])


class ParseResultBytes(_BytesNetloc):
    _fields = _obj(("scheme", "netloc", "path", "params", "query", "fragment"))

    @property
    def scheme(self):
        return self._vals[0]

    @property
    def netloc(self):
        return self._vals[1]

    @property
    def path(self):
        return self._vals[2]

    @property
    def params(self):
        return self._vals[3]

    @property
    def query(self):
        return self._vals[4]

    @property
    def fragment(self):
        return self._vals[5]

    def geturl(self):
        return urlunparse(self)

    def _replace(self, **kwargs):
        d = self._asdict()
        d.update(kwargs)
        return type(self)(d["scheme"], d["netloc"], d["path"], d["params"], d["query"], d["fragment"])


def _decode_result(r, encoding, errors):
    vals = [None if v is None else v.decode(encoding, errors) for v in r._vals]
    if isinstance(r, DefragResult):
        return DefragResult(*vals)
    if isinstance(r, SplitResultBytes) or isinstance(r, SplitResult):
        return SplitResult(*vals)
    return ParseResult(*vals)


def _encode_result(r, encoding, errors):
    vals = [None if v is None else v.encode(encoding, errors) for v in r._vals]
    if isinstance(r, DefragResult):
        return DefragResultBytes(*vals)
    if isinstance(r, SplitResult):
        return SplitResultBytes(*vals)
    return ParseResultBytes(*vals)


def _noop(obj):
    return obj


def _coerce_args(*args):
    str_input = isinstance(args[0], str)
    for arg in args[1:]:
        if arg and isinstance(arg, str) != str_input:
            raise TypeError("Cannot mix str and non-str arguments")
    if str_input:
        return args + (_noop,)
    return _decode_args(args) + (_encode_result_any,)


def _decode_args(args, encoding="ascii", errors="strict"):
    return tuple(x.decode(encoding, errors) if x else "" for x in args)


def _encode_result_any(obj, encoding="ascii", errors="strict"):
    if isinstance(obj, str):
        return obj.encode(encoding, errors)
    return _encode_result(obj, encoding, errors)


def _splitparams(url):
    if "/" in url:
        i = url.find(";", url.rfind("/"))
        if i < 0:
            return url, ""
    else:
        i = url.find(";")
    return url[:i], url[i + 1:]


def _splitnetloc(url, start=0):
    delim = len(url)
    for c in "/?#":
        wdelim = url.find(c, start)
        if wdelim >= 0:
            delim = min(delim, wdelim)
    return url[start:delim], url[delim:]


def _checknetloc(netloc):
    if not netloc or netloc.isascii():
        return
    import unicodedata
    n = netloc.replace("@", "")
    n = n.replace(":", "")
    n = n.replace("#", "")
    n = n.replace("?", "")
    netloc2 = unicodedata.normalize("NFKC", n)
    if n == netloc2:
        return
    for c in "/?#@:":
        if c in netloc2:
            raise ValueError("netloc '" + netloc + "' contains invalid " + "characters under NFKC normalization")


def _check_bracketed_netloc(netloc):
    hostname_and_port = netloc.rpartition("@")[2]
    before_bracket, have_open_br, bracketed = hostname_and_port.partition("[")
    if have_open_br:
        if before_bracket:
            raise ValueError("Invalid IPv6 URL")
        hostname, _, port = bracketed.partition("]")
        if port and not port.startswith(":"):
            raise ValueError("Invalid IPv6 URL")
    else:
        hostname, _, port = hostname_and_port.partition(":")
    _check_bracketed_host(hostname)


def _check_bracketed_host(hostname):
    if hostname.startswith(("v", "V")):
        if not re.match(r"\A[vV][a-fA-F0-9]+\..+\Z", hostname):
            raise ValueError("IPvFuture address is invalid")
    else:
        ip = ipaddress.ip_address(hostname)
        if isinstance(ip, ipaddress.IPv4Address):
            raise ValueError("An IPv4 address cannot be in brackets")


def urlparse(url, scheme="", allow_fragments=True):
    url, scheme, _coerce_result = _coerce_args(url, scheme)
    splitresult = urlsplit(url, scheme, allow_fragments)
    scheme, netloc, url, query, fragment = splitresult
    if scheme in uses_params and ";" in url:
        url, params = _splitparams(url)
    else:
        params = ""
    result = ParseResult(scheme, netloc, url, params, query, fragment)
    return _coerce_result(result)


def urlsplit(url, scheme="", allow_fragments=True):
    url, scheme, _coerce_result = _coerce_args(url, scheme)
    url = url.lstrip(_WHATWG_C0_CONTROL_OR_SPACE)
    scheme = scheme.strip(_WHATWG_C0_CONTROL_OR_SPACE)
    for b in _UNSAFE_URL_BYTES_TO_REMOVE:
        url = url.replace(b, "")
        scheme = scheme.replace(b, "")
    allow_fragments = bool(allow_fragments)
    netloc = query = fragment = ""
    i = url.find(":")
    if i > 0 and url[0].isascii() and url[0].isalpha():
        ok = True
        for c in url[:i]:
            if c not in scheme_chars:
                ok = False
                break
        if ok:
            scheme = url[:i].lower()
            url = url[i + 1:]
    if url[:2] == "//":
        netloc, url = _splitnetloc(url, 2)
        if ("[" in netloc and "]" not in netloc) or ("]" in netloc and "[" not in netloc):
            raise ValueError("Invalid IPv6 URL")
        if "[" in netloc and "]" in netloc:
            _check_bracketed_netloc(netloc)
    if allow_fragments and "#" in url:
        url, fragment = url.split("#", 1)
    if "?" in url:
        url, query = url.split("?", 1)
    _checknetloc(netloc)
    v = SplitResult(scheme, netloc, url, query, fragment)
    return _coerce_result(v)


def urlunparse(components):
    scheme, netloc, url, params, query, fragment, _coerce_result = _coerce_args(*components)
    if params:
        url = "%s;%s" % (url, params)
    return _coerce_result(urlunsplit((scheme, netloc, url, query, fragment)))


def urlunsplit(components):
    scheme, netloc, url, query, fragment, _coerce_result = _coerce_args(*components)
    if netloc:
        if url and url[:1] != "/":
            url = "/" + url
        url = "//" + netloc + url
    elif url[:2] == "//":
        url = "//" + url
    elif scheme and scheme in uses_netloc and (not url or url[:1] == "/"):
        url = "//" + url
    if scheme:
        url = scheme + ":" + url
    if query:
        url = url + "?" + query
    if fragment:
        url = url + "#" + fragment
    return _coerce_result(url)


def urljoin(base, url, allow_fragments=True):
    if not base:
        return url
    if not url:
        return base
    base, url, _coerce_result = _coerce_args(base, url)
    bscheme, bnetloc, bpath, bparams, bquery, bfragment = urlparse(base, "", allow_fragments)
    scheme, netloc, path, params, query, fragment = urlparse(url, bscheme, allow_fragments)
    if scheme != bscheme or scheme not in uses_relative:
        return _coerce_result(url)
    if scheme in uses_netloc:
        if netloc:
            return _coerce_result(urlunparse((scheme, netloc, path, params, query, fragment)))
        netloc = bnetloc
    if not path and not params:
        path = bpath
        params = bparams
        if not query:
            query = bquery
        return _coerce_result(urlunparse((scheme, netloc, path, params, query, fragment)))
    base_parts = bpath.split("/")
    if base_parts[-1] != "":
        del base_parts[-1]
    if path[:1] == "/":
        segments = path.split("/")
    else:
        segments = base_parts + path.split("/")
        segments[1:-1] = [s for s in segments[1:-1] if s]
    resolved_path = []
    for seg in segments:
        if seg == "..":
            if resolved_path:
                resolved_path.pop()
        elif seg == ".":
            continue
        else:
            resolved_path.append(seg)
    if segments[-1] in (".", ".."):
        resolved_path.append("")
    return _coerce_result(urlunparse((scheme, netloc, "/".join(resolved_path) or "/", params, query, fragment)))


def urldefrag(url):
    url, _coerce_result = _coerce_args(url)
    if "#" in url:
        s, n, p, a, q, frag = urlparse(url)
        defrag = urlunparse((s, n, p, a, q, ""))
    else:
        frag = ""
        defrag = url
    return _coerce_result(DefragResult(defrag, frag))


_hexdig = "0123456789ABCDEFabcdef"
_hextobyte = None


def unquote_to_bytes(string):
    if not string:
        return b""
    if isinstance(string, str):
        string = string.encode("utf-8")
    bits = string.split(b"%")
    if len(bits) == 1:
        return string
    res = [bits[0]]
    for item in bits[1:]:
        h = item[:2]
        if len(h) == 2 and h[0:1].decode("latin-1") in _hexdig and h[1:2].decode("latin-1") in _hexdig:
            res.append(bytes([int(h.decode("ascii"), 16)]))
            res.append(item[2:])
        else:
            res.append(b"%")
            res.append(item)
    return b"".join(res)


def unquote(string, encoding="utf-8", errors="replace"):
    if isinstance(string, bytes):
        return unquote_to_bytes(string).decode(encoding, errors)
    if "%" not in string:
        return string
    if encoding is None:
        encoding = "utf-8"
    if errors is None:
        errors = "replace"
    out = []
    pending = bytearray()
    i = 0
    n = len(string)
    while i < n:
        c = string[i]
        if c == "%" and i + 2 < n + 0 + 1 and i + 2 <= n - 1 + 0 and string[i + 1] in _hexdig and string[i + 2] in _hexdig:
            pending.append(int(string[i + 1:i + 3], 16))
            i += 3
            continue
        if pending:
            out.append(pending.decode(encoding, errors))
            pending = bytearray()
        out.append(c)
        i += 1
    if pending:
        out.append(pending.decode(encoding, errors))
    return "".join(out)


def parse_qs(qs, keep_blank_values=False, strict_parsing=False, encoding="utf-8", errors="replace",
             max_num_fields=None, separator="&"):
    parsed_result = {}
    pairs = parse_qsl(qs, keep_blank_values, strict_parsing, encoding=encoding, errors=errors,
                      max_num_fields=max_num_fields, separator=separator)
    for name, value in pairs:
        if name in parsed_result:
            parsed_result[name].append(value)
        else:
            parsed_result[name] = [value]
    return parsed_result


def parse_qsl(qs, keep_blank_values=False, strict_parsing=False, encoding="utf-8", errors="replace",
              max_num_fields=None, separator="&"):
    if not separator or not isinstance(separator, (str, bytes)):
        raise ValueError("Separator must be of type string or bytes.")
    if isinstance(qs, str):
        if not isinstance(separator, str):
            separator = str(separator, "ascii")
        eq = "="
        amp = "&" if False else separator
        qs_str = qs
    else:
        qs_str = None
        if qs is None:
            qs = b""
        if max_num_fields is not None:
            num_fields = 1 + qs.count(separator if isinstance(separator, bytes) else separator.encode("ascii"))
            if max_num_fields < num_fields:
                raise ValueError("Max number of fields exceeded")
        sep = separator if isinstance(separator, bytes) else separator.encode("ascii")
        r = []
        for name_value in qs.split(sep):
            if not name_value and not strict_parsing:
                continue
            nv = name_value.split(b"=", 1)
            if len(nv) != 2:
                if strict_parsing:
                    raise ValueError("bad query field: %r" % (name_value,))
                if keep_blank_values:
                    nv.append(b"")
                else:
                    continue
            if len(nv[1]) or keep_blank_values:
                name = nv[0].replace(b"+", b" ")
                name = unquote(name, encoding=encoding, errors=errors).encode(encoding, errors)
                value = nv[1].replace(b"+", b" ")
                value = unquote(value, encoding=encoding, errors=errors).encode(encoding, errors)
                r.append((name, value))
        return r
    if max_num_fields is not None:
        num_fields = 1 + qs_str.count(separator)
        if max_num_fields < num_fields:
            raise ValueError("Max number of fields exceeded")
    r = []
    for name_value in qs_str.split(separator):
        if not name_value and not strict_parsing:
            continue
        nv = name_value.split("=", 1)
        if len(nv) != 2:
            if strict_parsing:
                raise ValueError("bad query field: %r" % (name_value,))
            if keep_blank_values:
                nv.append("")
            else:
                continue
        if len(nv[1]) or keep_blank_values:
            name = nv[0].replace("+", " ")
            name = unquote(name, encoding=encoding, errors=errors)
            value = nv[1].replace("+", " ")
            value = unquote(value, encoding=encoding, errors=errors)
            r.append((name, value))
    return r


def unquote_plus(string, encoding="utf-8", errors="replace"):
    string = string.replace("+", " ")
    return unquote(string, encoding, errors)


_ALWAYS_SAFE = frozenset(b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_.-~")
_ALWAYS_SAFE_BYTES = bytes(_ALWAYS_SAFE)


def quote(string, safe="/", encoding=None, errors=None):
    if isinstance(string, str):
        if not string:
            return string
        if encoding is None:
            encoding = "utf-8"
        if errors is None:
            errors = "strict"
        string = string.encode(encoding, errors)
    else:
        if encoding is not None:
            raise TypeError("quote() doesn't support 'encoding' for bytes")
        if errors is not None:
            raise TypeError("quote() doesn't support 'errors' for bytes")
    return quote_from_bytes(string, safe)


def quote_plus(string, safe="", encoding=None, errors=None):
    if ((isinstance(string, str) and " " not in string) or (isinstance(string, bytes) and b" " not in string)):
        return quote(string, safe, encoding, errors)
    if isinstance(safe, str):
        space = " "
    else:
        space = b" "
    string = quote(string, safe + space, encoding, errors)
    return string.replace(" ", "+")


def quote_from_bytes(bs, safe="/"):
    if not isinstance(bs, (bytes, bytearray)):
        raise TypeError("quote_from_bytes() expected bytes")
    if not bs:
        return ""
    if isinstance(safe, str):
        safe = safe.encode("ascii", "ignore")
    else:
        safe = bytes([c for c in safe if c < 128])
    keep = set(_ALWAYS_SAFE)
    for c in safe:
        keep.add(c)
    out = []
    for c in bytes(bs):
        if c in keep:
            out.append(chr(c))
        else:
            out.append("%%%02X" % c)
    return "".join(out)


def urlencode(query, doseq=False, safe="", encoding=None, errors=None, quote_via=quote_plus):
    if hasattr(query, "items"):
        query = query.items()
    else:
        try:
            if len(query) and not isinstance(query[0], tuple):
                raise TypeError
        except TypeError as err:
            raise TypeError("not a valid non-string sequence or mapping object") from err
    l = []
    if not doseq:
        for k, v in query:
            if isinstance(k, bytes):
                k = quote_via(k, safe)
            else:
                k = quote_via(str(k), safe, encoding, errors)
            if isinstance(v, bytes):
                v = quote_via(v, safe)
            else:
                v = quote_via(str(v), safe, encoding, errors)
            l.append(k + "=" + v)
    else:
        for k, v in query:
            if isinstance(k, bytes):
                k = quote_via(k, safe)
            else:
                k = quote_via(str(k), safe, encoding, errors)
            if isinstance(v, bytes):
                v = quote_via(v, safe)
                l.append(k + "=" + v)
            elif isinstance(v, str):
                v = quote_via(v, safe, encoding, errors)
                l.append(k + "=" + v)
            else:
                try:
                    x = len(v)
                except TypeError:
                    v = quote_via(str(v), safe, encoding, errors)
                    l.append(k + "=" + v)
                else:
                    for elt in v:
                        if isinstance(elt, bytes):
                            elt = quote_via(elt, safe)
                        else:
                            elt = quote_via(str(elt), safe, encoding, errors)
                        l.append(k + "=" + elt)
    return "&".join(l)


def to_bytes(url):
    return str(url)


def unwrap(url):
    url = str(url).strip()
    if url[:1] == "<" and url[-1:] == ">":
        url = url[1:-1].strip()
    if url[:4] == "URL:":
        url = url[4:].strip()
    return url


def splittype(url):
    i = url.find(":")
    if i > 0 and re.fullmatch("[a-zA-Z][a-zA-Z0-9+.-]*", url[:i]):
        return url[:i].lower(), url[i + 1:]
    return None, url


def splithost(url):
    m = re.fullmatch("//([^/#?]*)(.*)", url, re.DOTALL)
    if m:
        host_port, path = m.groups()
        if path and path[0] != "/":
            path = "/" + path
        return host_port, path
    return None, url


def splituser(host):
    user, delim, host = host.rpartition("@")
    return (user if delim else None), host


def splitpasswd(user):
    user, delim, passwd = user.partition(":")
    return user, (passwd if delim else None)


def splitport(host):
    m = re.fullmatch("(.*):([0-9]*)", host, re.DOTALL)
    if m:
        host, port = m.groups()
        if port:
            return host, port
    return host, None


def splitnport(host, defport=-1):
    host, delim, port = host.rpartition(":")
    if not delim:
        host = port
    elif port:
        if port.isdigit() and port.isascii():
            nport = int(port)
        else:
            nport = None
        return host, nport
    return host, defport


def splitquery(url):
    path, delim, query = url.rpartition("?")
    if delim:
        return path, query
    return url, None


def splittag(url):
    path, delim, tag = url.rpartition("#")
    if delim:
        return path, tag
    return url, None


def splitattr(url):
    words = url.split(";")
    return words[0], words[1:]


def splitvalue(attr):
    attr, delim, value = attr.partition("=")
    return attr, (value if delim else None)
