"""Exception classes raised by urllib.

The base exception class is URLError, which inherits from OSError.  It
doesn't define any behavior of its own, but is the base class for all
exceptions defined in this package.

HTTPError is an exception class that is also a valid HTTP response
instance.  It behaves this way because HTTP protocol errors are valid
responses, with a status code, headers, and a body.  In some contexts,
an application may want to handle an exception like a regular
response.
"""
import io


def _obj(x: object) -> object:
    return x

__all__ = ['URLError', 'HTTPError', 'ContentTooShortError']

class URLError(OSError):

    def __init__(self, reason, filename=None):
        self.args = (reason,)
        self.reason = _obj(reason)
        if filename is not None:
            self.filename = _obj(filename)

    def __str__(self):
        return '<urlopen error %s>' % self.reason

class HTTPError(URLError):
    """Raised when HTTP error occurs, but also acts like non-error return"""

    def __init__(self, url, code, msg, hdrs, fp):
        self.code = _obj(code)
        self.msg = _obj(msg)
        self.hdrs = _obj(hdrs)
        self.filename = _obj(url)
        if fp is None:
            fp = io.BytesIO()
        self.fp = _obj(fp)
        self.args = (msg,)
        self.reason = _obj(msg)
        self.headers = _obj(hdrs)

    def __str__(self):
        return 'HTTP Error %s: %s' % (self.code, self.msg)

    def __repr__(self):
        return '<HTTPError %s: %r>' % (self.code, self.msg)

    def read(self, n=-1):
        return self.fp.read(n)

    def readline(self, limit=-1):
        return self.fp.readline(limit)

    def info(self):
        return self.hdrs

    def geturl(self):
        return self.filename

    def getcode(self):
        return self.code

    @property
    def status(self):
        return self.code

    def close(self):
        self.fp.close()

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()


class ContentTooShortError(URLError):
    """Exception raised when downloaded size does not match content-length."""

    def __init__(self, message, content):
        self.args = (message,)
        self.reason = _obj(message)
        self.content = content
