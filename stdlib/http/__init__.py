"""http: HTTP status codes and methods (hand-written: the enums are plain classes here)."""

__all__ = ['HTTPStatus', 'HTTPMethod']

_STATUS_TABLE = {
    100: ('CONTINUE', 'Continue', 'Request received, please continue'),
    101: ('SWITCHING_PROTOCOLS', 'Switching Protocols', 'Switching to new protocol; obey Upgrade header'),
    102: ('PROCESSING', 'Processing', ''),
    103: ('EARLY_HINTS', 'Early Hints', ''),
    200: ('OK', 'OK', 'Request fulfilled, document follows'),
    201: ('CREATED', 'Created', 'Document created, URL follows'),
    202: ('ACCEPTED', 'Accepted', 'Request accepted, processing continues off-line'),
    203: ('NON_AUTHORITATIVE_INFORMATION', 'Non-Authoritative Information', 'Request fulfilled from cache'),
    204: ('NO_CONTENT', 'No Content', 'Request fulfilled, nothing follows'),
    205: ('RESET_CONTENT', 'Reset Content', 'Clear input form for further input'),
    206: ('PARTIAL_CONTENT', 'Partial Content', 'Partial content follows'),
    207: ('MULTI_STATUS', 'Multi-Status', ''),
    208: ('ALREADY_REPORTED', 'Already Reported', ''),
    226: ('IM_USED', 'IM Used', ''),
    300: ('MULTIPLE_CHOICES', 'Multiple Choices', 'Object has several resources -- see URI list'),
    301: ('MOVED_PERMANENTLY', 'Moved Permanently', 'Object moved permanently -- see URI list'),
    302: ('FOUND', 'Found', 'Object moved temporarily -- see URI list'),
    303: ('SEE_OTHER', 'See Other', 'Object moved -- see Method and URL list'),
    304: ('NOT_MODIFIED', 'Not Modified', 'Document has not changed since given time'),
    305: ('USE_PROXY', 'Use Proxy', 'You must use proxy specified in Location to access this resource'),
    307: ('TEMPORARY_REDIRECT', 'Temporary Redirect', 'Object moved temporarily -- see URI list'),
    308: ('PERMANENT_REDIRECT', 'Permanent Redirect', 'Object moved permanently -- see URI list'),
    400: ('BAD_REQUEST', 'Bad Request', 'Bad request syntax or unsupported method'),
    401: ('UNAUTHORIZED', 'Unauthorized', 'No permission -- see authorization schemes'),
    402: ('PAYMENT_REQUIRED', 'Payment Required', 'No payment -- see charging schemes'),
    403: ('FORBIDDEN', 'Forbidden', 'Request forbidden -- authorization will not help'),
    404: ('NOT_FOUND', 'Not Found', 'Nothing matches the given URI'),
    405: ('METHOD_NOT_ALLOWED', 'Method Not Allowed', 'Specified method is invalid for this resource'),
    406: ('NOT_ACCEPTABLE', 'Not Acceptable', 'URI not available in preferred format'),
    407: ('PROXY_AUTHENTICATION_REQUIRED', 'Proxy Authentication Required', 'You must authenticate with this proxy before proceeding'),
    408: ('REQUEST_TIMEOUT', 'Request Timeout', 'Request timed out; try again later'),
    409: ('CONFLICT', 'Conflict', 'Request conflict'),
    410: ('GONE', 'Gone', 'URI no longer exists and has been permanently removed'),
    411: ('LENGTH_REQUIRED', 'Length Required', 'Client must specify Content-Length'),
    412: ('PRECONDITION_FAILED', 'Precondition Failed', 'Precondition in headers is false'),
    413: ('CONTENT_TOO_LARGE', 'Content Too Large', 'Content is too large'),
    414: ('URI_TOO_LONG', 'URI Too Long', 'URI is too long'),
    415: ('UNSUPPORTED_MEDIA_TYPE', 'Unsupported Media Type', 'Entity body in unsupported format'),
    416: ('RANGE_NOT_SATISFIABLE', 'Range Not Satisfiable', 'Cannot satisfy request range'),
    417: ('EXPECTATION_FAILED', 'Expectation Failed', 'Expect condition could not be satisfied'),
    418: ('IM_A_TEAPOT', "I'm a Teapot", 'Server refuses to brew coffee because it is a teapot.'),
    421: ('MISDIRECTED_REQUEST', 'Misdirected Request', 'Server is not able to produce a response'),
    422: ('UNPROCESSABLE_CONTENT', 'Unprocessable Content', ''),
    423: ('LOCKED', 'Locked', ''),
    424: ('FAILED_DEPENDENCY', 'Failed Dependency', ''),
    425: ('TOO_EARLY', 'Too Early', ''),
    426: ('UPGRADE_REQUIRED', 'Upgrade Required', ''),
    428: ('PRECONDITION_REQUIRED', 'Precondition Required', 'The origin server requires the request to be conditional'),
    429: ('TOO_MANY_REQUESTS', 'Too Many Requests', 'The user has sent too many requests in a given amount of time ("rate limiting")'),
    431: ('REQUEST_HEADER_FIELDS_TOO_LARGE', 'Request Header Fields Too Large', 'The server is unwilling to process the request because its header fields are too large'),
    451: ('UNAVAILABLE_FOR_LEGAL_REASONS', 'Unavailable For Legal Reasons', 'The server is denying access to the resource as a consequence of a legal demand'),
    500: ('INTERNAL_SERVER_ERROR', 'Internal Server Error', 'Server got itself in trouble'),
    501: ('NOT_IMPLEMENTED', 'Not Implemented', 'Server does not support this operation'),
    502: ('BAD_GATEWAY', 'Bad Gateway', 'Invalid responses from another server/proxy'),
    503: ('SERVICE_UNAVAILABLE', 'Service Unavailable', 'The server cannot process the request due to a high load'),
    504: ('GATEWAY_TIMEOUT', 'Gateway Timeout', 'The gateway server did not receive a timely response'),
    505: ('HTTP_VERSION_NOT_SUPPORTED', 'HTTP Version Not Supported', 'Cannot fulfill request'),
    506: ('VARIANT_ALSO_NEGOTIATES', 'Variant Also Negotiates', ''),
    507: ('INSUFFICIENT_STORAGE', 'Insufficient Storage', ''),
    508: ('LOOP_DETECTED', 'Loop Detected', ''),
    510: ('NOT_EXTENDED', 'Not Extended', ''),
    511: ('NETWORK_AUTHENTICATION_REQUIRED', 'Network Authentication Required', 'The client needs to authenticate to gain network access'),
}


class HTTPStatus:
    def __init__(self, value, name=None, phrase=None, description=None):
        if isinstance(value, HTTPStatus):
            value = value.value
        if name is None:
            if value not in _STATUS_TABLE:
                raise ValueError("%r is not a valid HTTPStatus" % (value,))
            info = _STATUS_TABLE[value]
            name = info[0]
            phrase = info[1]
            description = info[2]
        self.value = value
        self.name = name
        self.phrase = phrase
        self.description = description
        self._value_ = value
        self._name_ = name

    @property
    def is_informational(self):
        return 100 <= self.value <= 199

    @property
    def is_success(self):
        return 200 <= self.value <= 299

    @property
    def is_redirection(self):
        return 300 <= self.value <= 399

    @property
    def is_client_error(self):
        return 400 <= self.value <= 499

    @property
    def is_server_error(self):
        return 500 <= self.value <= 599

    def __int__(self):
        return self.value

    def __index__(self):
        return self.value

    def __eq__(self, other):
        if isinstance(other, HTTPStatus):
            return self.value == other.value
        return self.value == other

    def __ne__(self, other):
        return not self == other

    def __lt__(self, other):
        if isinstance(other, HTTPStatus):
            return self.value < other.value
        return self.value < other

    def __le__(self, other):
        return self < other or self == other

    def __gt__(self, other):
        return not (self < other or self == other)

    def __ge__(self, other):
        return not self < other

    def __hash__(self):
        return hash(self.value)

    def __str__(self):
        return str(self.value)

    def __repr__(self):
        return "<HTTPStatus.%s: %d>" % (self.name, self.value)

    def __format__(self, spec):
        return format(self.value, spec)


HTTPStatus.CONTINUE = HTTPStatus(100)
HTTPStatus.SWITCHING_PROTOCOLS = HTTPStatus(101)
HTTPStatus.PROCESSING = HTTPStatus(102)
HTTPStatus.EARLY_HINTS = HTTPStatus(103)
HTTPStatus.OK = HTTPStatus(200)
HTTPStatus.CREATED = HTTPStatus(201)
HTTPStatus.ACCEPTED = HTTPStatus(202)
HTTPStatus.NON_AUTHORITATIVE_INFORMATION = HTTPStatus(203)
HTTPStatus.NO_CONTENT = HTTPStatus(204)
HTTPStatus.RESET_CONTENT = HTTPStatus(205)
HTTPStatus.PARTIAL_CONTENT = HTTPStatus(206)
HTTPStatus.MULTI_STATUS = HTTPStatus(207)
HTTPStatus.ALREADY_REPORTED = HTTPStatus(208)
HTTPStatus.IM_USED = HTTPStatus(226)
HTTPStatus.MULTIPLE_CHOICES = HTTPStatus(300)
HTTPStatus.MOVED_PERMANENTLY = HTTPStatus(301)
HTTPStatus.FOUND = HTTPStatus(302)
HTTPStatus.SEE_OTHER = HTTPStatus(303)
HTTPStatus.NOT_MODIFIED = HTTPStatus(304)
HTTPStatus.USE_PROXY = HTTPStatus(305)
HTTPStatus.TEMPORARY_REDIRECT = HTTPStatus(307)
HTTPStatus.PERMANENT_REDIRECT = HTTPStatus(308)
HTTPStatus.BAD_REQUEST = HTTPStatus(400)
HTTPStatus.UNAUTHORIZED = HTTPStatus(401)
HTTPStatus.PAYMENT_REQUIRED = HTTPStatus(402)
HTTPStatus.FORBIDDEN = HTTPStatus(403)
HTTPStatus.NOT_FOUND = HTTPStatus(404)
HTTPStatus.METHOD_NOT_ALLOWED = HTTPStatus(405)
HTTPStatus.NOT_ACCEPTABLE = HTTPStatus(406)
HTTPStatus.PROXY_AUTHENTICATION_REQUIRED = HTTPStatus(407)
HTTPStatus.REQUEST_TIMEOUT = HTTPStatus(408)
HTTPStatus.CONFLICT = HTTPStatus(409)
HTTPStatus.GONE = HTTPStatus(410)
HTTPStatus.LENGTH_REQUIRED = HTTPStatus(411)
HTTPStatus.PRECONDITION_FAILED = HTTPStatus(412)
HTTPStatus.CONTENT_TOO_LARGE = HTTPStatus(413)
HTTPStatus.URI_TOO_LONG = HTTPStatus(414)
HTTPStatus.UNSUPPORTED_MEDIA_TYPE = HTTPStatus(415)
HTTPStatus.RANGE_NOT_SATISFIABLE = HTTPStatus(416)
HTTPStatus.EXPECTATION_FAILED = HTTPStatus(417)
HTTPStatus.IM_A_TEAPOT = HTTPStatus(418)
HTTPStatus.MISDIRECTED_REQUEST = HTTPStatus(421)
HTTPStatus.UNPROCESSABLE_CONTENT = HTTPStatus(422)
HTTPStatus.LOCKED = HTTPStatus(423)
HTTPStatus.FAILED_DEPENDENCY = HTTPStatus(424)
HTTPStatus.TOO_EARLY = HTTPStatus(425)
HTTPStatus.UPGRADE_REQUIRED = HTTPStatus(426)
HTTPStatus.PRECONDITION_REQUIRED = HTTPStatus(428)
HTTPStatus.TOO_MANY_REQUESTS = HTTPStatus(429)
HTTPStatus.REQUEST_HEADER_FIELDS_TOO_LARGE = HTTPStatus(431)
HTTPStatus.UNAVAILABLE_FOR_LEGAL_REASONS = HTTPStatus(451)
HTTPStatus.INTERNAL_SERVER_ERROR = HTTPStatus(500)
HTTPStatus.NOT_IMPLEMENTED = HTTPStatus(501)
HTTPStatus.BAD_GATEWAY = HTTPStatus(502)
HTTPStatus.SERVICE_UNAVAILABLE = HTTPStatus(503)
HTTPStatus.GATEWAY_TIMEOUT = HTTPStatus(504)
HTTPStatus.HTTP_VERSION_NOT_SUPPORTED = HTTPStatus(505)
HTTPStatus.VARIANT_ALSO_NEGOTIATES = HTTPStatus(506)
HTTPStatus.INSUFFICIENT_STORAGE = HTTPStatus(507)
HTTPStatus.LOOP_DETECTED = HTTPStatus(508)
HTTPStatus.NOT_EXTENDED = HTTPStatus(510)
HTTPStatus.NETWORK_AUTHENTICATION_REQUIRED = HTTPStatus(511)


class HTTPMethod:
    def __init__(self, value, description=''):
        self.value = value
        self.name = value
        self.description = description

    def __eq__(self, other):
        if isinstance(other, HTTPMethod):
            return self.value == other.value
        return self.value == other

    def __hash__(self):
        return hash(self.value)

    def __str__(self):
        return self.value

    def __repr__(self):
        return "<HTTPMethod.%s>" % self.name

HTTPMethod.CONNECT = HTTPMethod('CONNECT', 'Establish a connection to the server.')
HTTPMethod.DELETE = HTTPMethod('DELETE', 'Remove the target.')
HTTPMethod.GET = HTTPMethod('GET', 'Retrieve the target.')
HTTPMethod.HEAD = HTTPMethod('HEAD', 'Same as GET, but only retrieve the status line and header section.')
HTTPMethod.OPTIONS = HTTPMethod('OPTIONS', 'Describe the communication options for the target.')
HTTPMethod.PATCH = HTTPMethod('PATCH', 'Apply partial modifications to a target.')
HTTPMethod.POST = HTTPMethod('POST', 'Perform target-specific processing with the request payload.')
HTTPMethod.PUT = HTTPMethod('PUT', 'Replace the target with the request payload.')
HTTPMethod.TRACE = HTTPMethod('TRACE', 'Perform a message loop-back test along the path to the target.')
HTTPStatus.REQUEST_ENTITY_TOO_LARGE = HTTPStatus(413)
HTTPStatus.REQUEST_URI_TOO_LONG = HTTPStatus(414)
HTTPStatus.REQUESTED_RANGE_NOT_SATISFIABLE = HTTPStatus(416)
HTTPStatus.UNPROCESSABLE_ENTITY = HTTPStatus(422)
