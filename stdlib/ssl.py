"""ssl: TLS for sockets, over Go's crypto/tls (runtime/sslnative.go).

The API is CPython's for contexts and wrapped sockets: SSLContext (client
and server, certificates, verification, hostname checking, ALPN, versions),
SSLSocket (handshake, read, write, makefile, getpeercert, version, cipher,
unwrap), create_default_context, the error classes with OpenSSL's reasons
and verify codes, and the PEM/DER helpers. There is no memory BIO
(SSLObject, MemoryBIO), so asyncio cannot use TLS, no session reuse and no
post-handshake authentication. Cipher names are OpenSSL's, but Go chooses
the cipher, so it can differ from what OpenSSL would choose."""

import base64
import enum
import errno
import os
import socket as _socket
from socket import socket, SOCK_STREAM, create_connection, AF_INET

OPENSSL_VERSION = "Go crypto/tls (py2go)"
OPENSSL_VERSION_INFO = (3, 0, 0, 0, 0)
OPENSSL_VERSION_NUMBER = 0x30000000
_OPENSSL_API_VERSION = (3, 0, 0, 0, 0)

HAS_SNI = True
HAS_ECDH = True
HAS_NPN = False
HAS_ALPN = True
HAS_SSLv2 = False
HAS_SSLv3 = False
HAS_TLSv1 = True
HAS_TLSv1_1 = True
HAS_TLSv1_2 = True
HAS_TLSv1_3 = True
HAS_NEVER_CHECK_COMMON_NAME = True
HAS_PSK = False
HAS_PHA = False
CHANNEL_BINDING_TYPES = ['tls-unique']

PEM_HEADER = "-----BEGIN CERTIFICATE-----"
PEM_FOOTER = "-----END CERTIFICATE-----"

socket_error = OSError


class _SSLMethod(enum.IntEnum):
    PROTOCOL_SSLv23 = 2
    PROTOCOL_TLS_CLIENT = 16
    PROTOCOL_TLS_SERVER = 17


PROTOCOL_SSLv23 = _SSLMethod.PROTOCOL_SSLv23
PROTOCOL_TLS = _SSLMethod.PROTOCOL_SSLv23
PROTOCOL_TLS_CLIENT = _SSLMethod.PROTOCOL_TLS_CLIENT
PROTOCOL_TLS_SERVER = _SSLMethod.PROTOCOL_TLS_SERVER


class VerifyMode(enum.IntEnum):
    CERT_NONE = 0
    CERT_OPTIONAL = 1
    CERT_REQUIRED = 2


CERT_NONE = VerifyMode.CERT_NONE
CERT_OPTIONAL = VerifyMode.CERT_OPTIONAL
CERT_REQUIRED = VerifyMode.CERT_REQUIRED


class VerifyFlags(enum.IntFlag):
    VERIFY_DEFAULT = 0
    VERIFY_CRL_CHECK_LEAF = 4
    VERIFY_CRL_CHECK_CHAIN = 12
    VERIFY_X509_STRICT = 32
    VERIFY_ALLOW_PROXY_CERTS = 64
    VERIFY_X509_TRUSTED_FIRST = 32768
    VERIFY_X509_PARTIAL_CHAIN = 524288


VERIFY_DEFAULT = VerifyFlags.VERIFY_DEFAULT
VERIFY_CRL_CHECK_LEAF = VerifyFlags.VERIFY_CRL_CHECK_LEAF
VERIFY_CRL_CHECK_CHAIN = VerifyFlags.VERIFY_CRL_CHECK_CHAIN
VERIFY_X509_STRICT = VerifyFlags.VERIFY_X509_STRICT
VERIFY_ALLOW_PROXY_CERTS = VerifyFlags.VERIFY_ALLOW_PROXY_CERTS
VERIFY_X509_TRUSTED_FIRST = VerifyFlags.VERIFY_X509_TRUSTED_FIRST
VERIFY_X509_PARTIAL_CHAIN = VerifyFlags.VERIFY_X509_PARTIAL_CHAIN


class Options(enum.IntFlag):
    OP_ALL = 2147483728
    OP_NO_SSLv2 = 0
    OP_NO_SSLv3 = 33554432
    OP_NO_TLSv1 = 67108864
    OP_NO_TLSv1_1 = 268435456
    OP_NO_TLSv1_2 = 134217728
    OP_NO_TLSv1_3 = 536870912
    OP_CIPHER_SERVER_PREFERENCE = 4194304
    OP_SINGLE_DH_USE = 0
    OP_SINGLE_ECDH_USE = 0
    OP_NO_COMPRESSION = 131072
    OP_NO_TICKET = 16384
    OP_NO_RENEGOTIATION = 1073741824
    OP_ENABLE_MIDDLEBOX_COMPAT = 1048576
    OP_IGNORE_UNEXPECTED_EOF = 128
    OP_LEGACY_SERVER_CONNECT = 4


OP_ALL = Options.OP_ALL
OP_NO_SSLv2 = Options.OP_NO_SSLv2
OP_NO_SSLv3 = Options.OP_NO_SSLv3
OP_NO_TLSv1 = Options.OP_NO_TLSv1
OP_NO_TLSv1_1 = Options.OP_NO_TLSv1_1
OP_NO_TLSv1_2 = Options.OP_NO_TLSv1_2
OP_NO_TLSv1_3 = Options.OP_NO_TLSv1_3
OP_CIPHER_SERVER_PREFERENCE = Options.OP_CIPHER_SERVER_PREFERENCE
OP_SINGLE_DH_USE = Options.OP_SINGLE_DH_USE
OP_SINGLE_ECDH_USE = Options.OP_SINGLE_ECDH_USE
OP_NO_COMPRESSION = Options.OP_NO_COMPRESSION
OP_NO_TICKET = Options.OP_NO_TICKET
OP_NO_RENEGOTIATION = Options.OP_NO_RENEGOTIATION
OP_ENABLE_MIDDLEBOX_COMPAT = Options.OP_ENABLE_MIDDLEBOX_COMPAT
OP_IGNORE_UNEXPECTED_EOF = Options.OP_IGNORE_UNEXPECTED_EOF
OP_LEGACY_SERVER_CONNECT = Options.OP_LEGACY_SERVER_CONNECT


class TLSVersion(enum.IntEnum):
    MINIMUM_SUPPORTED = -2
    MAXIMUM_SUPPORTED = -1
    SSLv3 = 768
    TLSv1 = 769
    TLSv1_1 = 770
    TLSv1_2 = 771
    TLSv1_3 = 772


class SSLErrorNumber(enum.IntEnum):
    SSL_ERROR_SSL = 1
    SSL_ERROR_WANT_READ = 2
    SSL_ERROR_WANT_WRITE = 3
    SSL_ERROR_WANT_X509_LOOKUP = 4
    SSL_ERROR_SYSCALL = 5
    SSL_ERROR_ZERO_RETURN = 6
    SSL_ERROR_WANT_CONNECT = 7
    SSL_ERROR_EOF = 8
    SSL_ERROR_INVALID_ERROR_CODE = 10


SSL_ERROR_SSL = SSLErrorNumber.SSL_ERROR_SSL
SSL_ERROR_WANT_READ = SSLErrorNumber.SSL_ERROR_WANT_READ
SSL_ERROR_WANT_WRITE = SSLErrorNumber.SSL_ERROR_WANT_WRITE
SSL_ERROR_WANT_X509_LOOKUP = SSLErrorNumber.SSL_ERROR_WANT_X509_LOOKUP
SSL_ERROR_SYSCALL = SSLErrorNumber.SSL_ERROR_SYSCALL
SSL_ERROR_ZERO_RETURN = SSLErrorNumber.SSL_ERROR_ZERO_RETURN
SSL_ERROR_WANT_CONNECT = SSLErrorNumber.SSL_ERROR_WANT_CONNECT
SSL_ERROR_EOF = SSLErrorNumber.SSL_ERROR_EOF
SSL_ERROR_INVALID_ERROR_CODE = SSLErrorNumber.SSL_ERROR_INVALID_ERROR_CODE


class Purpose(enum.Enum):
    SERVER_AUTH = '1.3.6.1.5.5.7.3.1'
    CLIENT_AUTH = '1.3.6.1.5.5.7.3.2'


class SSLError(OSError):
    """An error of the TLS layer: library is "SSL", reason OpenSSL's name
    for it."""

    library = None
    reason = None

    def __str__(self):
        if len(self.args) >= 2 and isinstance(self.args[1], str):
            return self.args[1]
        return OSError.__str__(self)


class SSLZeroReturnError(SSLError):
    pass


class SSLWantReadError(SSLError):
    pass


class SSLWantWriteError(SSLError):
    pass


class SSLSyscallError(SSLError):
    pass


class SSLEOFError(SSLError):
    pass


class SSLCertVerificationError(SSLError, ValueError):
    verify_code = None
    verify_message = None


CertificateError = SSLCertVerificationError


def _error(err):
    """The exception for an error tuple of the native part."""
    kind, code, reason, message, verify_code, verify_message = err
    text = "[SSL: %s] %s (_ssl.c:1032)" % (reason, message)
    if kind == "SSLCertVerificationError":
        e = SSLCertVerificationError(code, text)
        e.verify_code = verify_code
        e.verify_message = verify_message
    elif kind == "SSLEOFError":
        e = SSLEOFError(code, text)
    elif kind == "SSLZeroReturnError":
        e = SSLZeroReturnError(code, text)
    else:
        e = SSLError(code, text)
    e.library = "SSL"
    e.reason = reason
    return e


class SSLContext:
    """A TLS configuration: certificates, verification and versions."""

    def __init__(self, protocol=None):
        if protocol is None:
            protocol = PROTOCOL_TLS
        self.protocol = _SSLMethod(protocol)
        self._ctx = _native("_ssl.context_new", protocol == PROTOCOL_TLS_SERVER)
        if protocol == PROTOCOL_TLS_CLIENT:
            self._verify_mode = CERT_REQUIRED
            self._check_hostname = True
        else:
            self._verify_mode = CERT_NONE
            self._check_hostname = False
            _native("_ssl.context_set", self._ctx, "check_hostname", False)
            _native("_ssl.context_set", self._ctx, "verify_mode", 0)
        self.options = Options.OP_ALL | Options.OP_NO_SSLv3 | Options.OP_NO_COMPRESSION | Options.OP_CIPHER_SERVER_PREFERENCE
        self.verify_flags = VerifyFlags.VERIFY_X509_TRUSTED_FIRST | VerifyFlags.VERIFY_X509_PARTIAL_CHAIN
        self._minimum_version = TLSVersion.MINIMUM_SUPPORTED
        self._maximum_version = TLSVersion.MAXIMUM_SUPPORTED
        self.post_handshake_auth = None
        self.sni_callback = None
        self.keylog_filename = None
        self.security_level = 2
        self.num_tickets = 2
        self.hostname_checks_common_name = False
        self._alpn = []

    @property
    def verify_mode(self):
        return self._verify_mode

    @verify_mode.setter
    def verify_mode(self, value):
        value = VerifyMode(value)
        if value == CERT_NONE and self._check_hostname:
            raise ValueError("Cannot set verify_mode to CERT_NONE when check_hostname is enabled.")
        self._verify_mode = value
        _native("_ssl.context_set", self._ctx, "verify_mode", int(value))

    @property
    def check_hostname(self):
        return self._check_hostname

    @check_hostname.setter
    def check_hostname(self, value):
        value = bool(value)
        if value and self._verify_mode == CERT_NONE:
            self.verify_mode = CERT_REQUIRED
        self._check_hostname = value
        _native("_ssl.context_set", self._ctx, "check_hostname", value)

    @property
    def minimum_version(self):
        return self._minimum_version

    @minimum_version.setter
    def minimum_version(self, value):
        self._minimum_version = TLSVersion(value)
        _native("_ssl.context_set", self._ctx, "minimum_version", int(value))

    @property
    def maximum_version(self):
        return self._maximum_version

    @maximum_version.setter
    def maximum_version(self, value):
        self._maximum_version = TLSVersion(value)
        _native("_ssl.context_set", self._ctx, "maximum_version", int(value))

    def load_cert_chain(self, certfile, keyfile=None, password=None):
        certfile = os.fspath(certfile)
        if keyfile is not None:
            keyfile = os.fspath(keyfile)
        err = _native("_ssl.load_cert_chain", self._ctx, certfile, keyfile, password)
        if err is not None:
            raise _error(err)

    def load_verify_locations(self, cafile=None, capath=None, cadata=None):
        if cafile is None and capath is None and cadata is None:
            raise TypeError("cafile, capath and cadata cannot be all omitted")
        if cafile is not None:
            cafile = os.fspath(cafile)
        if capath is not None:
            capath = os.fspath(capath)
        err = _native("_ssl.load_verify_locations", self._ctx, cafile, capath, cadata)
        if err is not None:
            raise _error(err)

    def load_default_certs(self, purpose=Purpose.SERVER_AUTH):
        if not isinstance(purpose, Purpose):
            raise TypeError(purpose)
        self.set_default_verify_paths()

    def set_default_verify_paths(self):
        _native("_ssl.load_default_certs", self._ctx)

    def get_ca_certs(self, binary_form=False):
        return _native("_ssl.ca_certs", self._ctx, binary_form)

    def cert_store_stats(self):
        n = len(self.get_ca_certs(True))
        return {'x509': n, 'crl': 0, 'x509_ca': n}

    def set_ciphers(self, ciphers):
        # Go chooses among its own cipher suites.
        if not isinstance(ciphers, str):
            raise TypeError("ciphers must be a str")

    def get_ciphers(self):
        return []

    def set_alpn_protocols(self, alpn_protocols):
        protos = []
        for p in alpn_protocols:
            b = p.encode('ascii') if isinstance(p, str) else bytes(p)
            if len(b) == 0 or len(b) > 255:
                raise SSLError('ALPN protocols must be 1 to 255 in length')
            protos.append(b.decode('ascii'))
        self._alpn = protos
        _native("_ssl.context_set", self._ctx, "alpn", protos)

    def set_npn_protocols(self, npn_protocols):
        raise NotImplementedError("The NPN extension requires OpenSSL 1.0.1 or later.")

    def set_ecdh_curve(self, name):
        pass

    def set_servername_callback(self, server_name_callback):
        self.sni_callback = server_name_callback

    def session_stats(self):
        return {'number': 0, 'connect': 0, 'connect_good': 0, 'connect_renegotiate': 0,
                'accept': 0, 'accept_good': 0, 'accept_renegotiate': 0, 'hits': 0,
                'misses': 0, 'timeouts': 0, 'cache_full': 0}

    def wrap_socket(self, sock, server_side=False, do_handshake_on_connect=True,
                    suppress_ragged_eofs=True, server_hostname=None, session=None):
        return SSLSocket(sock, self, server_side, do_handshake_on_connect,
                         suppress_ragged_eofs, server_hostname)

    def wrap_bio(self, incoming, outgoing, server_side=False, server_hostname=None, session=None):
        raise NotImplementedError("py2go's ssl has no memory BIO")


class SSLSocket(socket):
    """A socket whose data goes through TLS. Made by SSLContext.wrap_socket."""

    def __init__(self, sock, context, server_side, do_handshake_on_connect,
                 suppress_ragged_eofs, server_hostname):
        if sock.getsockopt(_socket.SOL_SOCKET, _socket.SO_TYPE) != SOCK_STREAM:
            raise NotImplementedError("only stream sockets are supported")
        if server_side:
            if server_hostname:
                raise ValueError("server_hostname can only be specified in client mode")
        elif context.check_hostname and not server_hostname:
            raise ValueError("check_hostname requires server_hostname")
        timeout = sock.gettimeout()
        socket.__init__(self, sock.family, sock.type, sock.proto, sock.fileno())
        sock.detach()
        self.settimeout(timeout)
        self._context = context
        self.server_side = server_side
        self.server_hostname = server_hostname
        self.do_handshake_on_connect = do_handshake_on_connect
        self.suppress_ragged_eofs = suppress_ragged_eofs
        self._sslobj = None
        self._handshaken = False
        connected = True
        try:
            self.getpeername()
        except OSError as e:
            if e.errno != errno.ENOTCONN:
                raise
            connected = False
        if connected:
            self._sslobj = _native("_ssl.wrap", context._ctx, self.fileno(), server_side, server_hostname)
            if do_handshake_on_connect:
                if timeout == 0.0:
                    raise ValueError("do_handshake_on_connect should not be specified for non-blocking sockets")
                self.do_handshake()

    @property
    def context(self):
        return self._context

    @context.setter
    def context(self, ctx):
        self._context = ctx

    def _check_connected(self):
        if self._sslobj is None:
            self.getpeername()

    def do_handshake(self, block=False):
        self._check_connected()
        err = _native("_ssl.handshake", self._sslobj, self.gettimeout())
        if err is not None:
            raise _error(err)
        self._handshaken = True

    def read(self, len=1024, buffer=None):
        if self._sslobj is None:
            raise ValueError("Read on closed or unwrapped SSL socket.")
        if buffer is not None and len == 0:
            return 0
        data, err = _native("_ssl.read", self._sslobj, len, self.gettimeout(), self.suppress_ragged_eofs)
        if err is not None:
            raise _error(err)
        if buffer is not None:
            n = _builtin_len(data)
            buffer[:n] = data
            return n
        return data

    def write(self, data):
        if self._sslobj is None:
            raise ValueError("Write on closed or unwrapped SSL socket.")
        n, err = _native("_ssl.write", self._sslobj, data, self.gettimeout())
        if err is not None:
            raise _error(err)
        return n

    def getpeercert(self, binary_form=False):
        self._check_connected()
        return _native("_ssl.peercert", self._sslobj, binary_form)

    def _info(self):
        if self._sslobj is None:
            return None
        return _native("_ssl.info", self._sslobj)

    def selected_alpn_protocol(self):
        info = self._info()
        if info is None:
            return None
        return info[3]

    def selected_npn_protocol(self):
        return None

    def cipher(self):
        info = self._info()
        if info is None:
            return None
        return (info[1], info[0], info[2])

    def shared_ciphers(self):
        return None

    def compression(self):
        return None

    def version(self):
        info = self._info()
        if info is None:
            return None
        return info[0]

    def pending(self):
        return 0

    def send(self, data, flags=0):
        if flags != 0:
            raise ValueError("non-zero flags not allowed in calls to send() on %s" % self.__class__)
        return self.write(data)

    def sendall(self, data, flags=0):
        if flags != 0:
            raise ValueError("non-zero flags not allowed in calls to sendall() on %s" % self.__class__)
        data = bytes(data)
        total = _builtin_len(data)
        sent = 0
        while sent < total:
            sent += self.write(data[sent:])

    def sendto(self, data, flags_or_addr, addr=None):
        raise ValueError("sendto not allowed on instances of %s" % self.__class__)

    def recv(self, buflen=1024, flags=0):
        if flags != 0:
            raise ValueError("non-zero flags not allowed in calls to recv() on %s" % self.__class__)
        return self.read(buflen)

    def recv_into(self, buffer, nbytes=None, flags=0):
        if nbytes is None or nbytes == 0:
            nbytes = _builtin_len(buffer)
        if flags != 0:
            raise ValueError("non-zero flags not allowed in calls to recv_into() on %s" % self.__class__)
        return self.read(nbytes, buffer)

    def recvfrom(self, buflen=1024, flags=0):
        raise ValueError("recvfrom not allowed on instances of %s" % self.__class__)

    def recvfrom_into(self, buffer, nbytes=None, flags=0):
        raise ValueError("recvfrom_into not allowed on instances of %s" % self.__class__)

    def recvmsg(self, bufsize, ancbufsize=0, flags=0):
        raise NotImplementedError("recvmsg not allowed on instances of %s" % self.__class__)

    def sendmsg(self, buffers, ancdata=(), flags=0, address=None):
        raise NotImplementedError("sendmsg not allowed on instances of %s" % self.__class__)

    def shutdown(self, how):
        self._free()
        socket.shutdown(self, how)

    def unwrap(self):
        if self._sslobj is None:
            raise ValueError("No SSL wrapper around " + str(self))
        err = _native("_ssl.unwrap", self._sslobj, self.gettimeout())
        self._free()
        if err is not None:
            raise _error(err)
        return self

    def verify_client_post_handshake(self):
        raise NotImplementedError("post-handshake authentication is not supported")

    def _free(self):
        if self._sslobj is not None:
            _native("_ssl.free", self._sslobj)
            self._sslobj = None

    def _real_close(self):
        self._free()
        socket._real_close(self)

    def connect(self, address):
        if self._sslobj is not None:
            raise ValueError("attempt to connect already-connected SSLSocket!")
        socket.connect(self, address)
        self._sslobj = _native("_ssl.wrap", self._context._ctx, self.fileno(), self.server_side, self.server_hostname)
        if self.do_handshake_on_connect:
            self.do_handshake()

    def connect_ex(self, address):
        try:
            self.connect(address)
        except OSError as e:
            if e.errno is None:
                raise
            return e.errno
        return 0

    def accept(self):
        newsock, addr = socket.accept(self)
        newsock = self._context.wrap_socket(newsock, do_handshake_on_connect=self.do_handshake_on_connect,
                                            suppress_ragged_eofs=self.suppress_ragged_eofs, server_side=True)
        return newsock, addr

    def get_channel_binding(self, cb_type="tls-unique"):
        if cb_type not in CHANNEL_BINDING_TYPES:
            raise ValueError("Unsupported channel binding type")
        return None

    def dup(self):
        raise NotImplementedError("Can't dup() %s instances" % self.__class__.__name__)


def _builtin_len(x):
    # len() inside SSLSocket.read, whose parameter is named len.
    return len(x)


def create_default_context(purpose=Purpose.SERVER_AUTH, *, cafile=None, capath=None, cadata=None):
    if not isinstance(purpose, Purpose):
        raise TypeError(purpose)
    if purpose == Purpose.SERVER_AUTH:
        context = SSLContext(PROTOCOL_TLS_CLIENT)
        context.verify_mode = CERT_REQUIRED
        context.check_hostname = True
    else:
        context = SSLContext(PROTOCOL_TLS_SERVER)
    if cafile or capath or cadata:
        context.load_verify_locations(cafile, capath, cadata)
    elif context.verify_mode != CERT_NONE:
        context.load_default_certs(purpose)
    return context


def _create_unverified_context(protocol=None, *, cert_reqs=CERT_NONE, check_hostname=False,
                               purpose=Purpose.SERVER_AUTH, certfile=None, keyfile=None,
                               cafile=None, capath=None, cadata=None):
    if not isinstance(purpose, Purpose):
        raise TypeError(purpose)
    if purpose == Purpose.SERVER_AUTH:
        if protocol is None:
            protocol = PROTOCOL_TLS_CLIENT
    elif purpose == Purpose.CLIENT_AUTH:
        if protocol is None:
            protocol = PROTOCOL_TLS_SERVER
    context = SSLContext(protocol)
    context.check_hostname = check_hostname
    if cert_reqs is not None:
        context.verify_mode = cert_reqs
    if check_hostname:
        context.check_hostname = True
    if keyfile and not certfile:
        raise ValueError("certfile must be specified")
    if certfile or keyfile:
        context.load_cert_chain(certfile, keyfile)
    if cafile or capath or cadata:
        context.load_verify_locations(cafile, capath, cadata)
    elif context.verify_mode != CERT_NONE:
        context.load_default_certs(purpose)
    return context


_create_default_https_context = create_default_context
_create_stdlib_context = _create_unverified_context


def DER_cert_to_PEM_cert(der_cert_bytes):
    """Takes a certificate in binary DER format and returns the PEM version
    of it as a string."""
    f = base64.standard_b64encode(der_cert_bytes).decode('ASCII')
    ss = [PEM_HEADER]
    for i in range(0, len(f), 64):
        ss.append(f[i:i + 64])
    ss.append(PEM_FOOTER + '\n')
    return '\n'.join(ss)


def PEM_cert_to_DER_cert(pem_cert_string):
    """Takes a certificate in ASCII PEM format and returns the DER-encoded
    version of it as a byte sequence."""
    if not pem_cert_string.startswith(PEM_HEADER):
        raise ValueError("Invalid PEM encoding; must start with %s" % PEM_HEADER)
    if not pem_cert_string.strip().endswith(PEM_FOOTER):
        raise ValueError("Invalid PEM encoding; must end with %s" % PEM_FOOTER)
    d = pem_cert_string.strip()[len(PEM_HEADER):-len(PEM_FOOTER)]
    return base64.decodebytes(d.encode('ASCII', 'strict'))


def get_server_certificate(addr, ssl_version=PROTOCOL_TLS_CLIENT, ca_certs=None, timeout=_socket._GLOBAL_DEFAULT_TIMEOUT):
    """Retrieve the certificate from the server at the specified address,
    and return it as a PEM-encoded string."""
    host, port = addr
    if ca_certs is not None:
        cert_reqs = CERT_REQUIRED
    else:
        cert_reqs = CERT_NONE
    context = _create_unverified_context(ssl_version, cert_reqs=cert_reqs, cafile=ca_certs)
    with create_connection(addr, timeout=timeout) as sock:
        with context.wrap_socket(sock, server_hostname=host) as sslsock:
            dercert = sslsock.getpeercert(True)
    return DER_cert_to_PEM_cert(dercert)


def cert_time_to_seconds(cert_time):
    """Return the time in seconds since the Epoch, given the timestring
    representing the "notBefore" or "notAfter" date from a certificate in
    ``"%b %d %H:%M:%S %Y %Z"`` strptime format (C locale)."""
    from time import strptime
    from calendar import timegm
    months = ("Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec")
    time_format = ' %d %H:%M:%S %Y GMT'
    try:
        month_number = months.index(cert_time[:3].title()) + 1
    except ValueError:
        raise ValueError('time data %r does not match format "%%b%s"' % (cert_time, time_format))
    else:
        tt = strptime(cert_time[3:], time_format)
        return timegm((tt[0], month_number) + tuple(tt[2:6]))


class DefaultVerifyPaths:
    def __init__(self, cafile, capath, openssl_cafile_env, openssl_cafile, openssl_capath_env, openssl_capath):
        self.cafile = cafile
        self.capath = capath
        self.openssl_cafile_env = openssl_cafile_env
        self.openssl_cafile = openssl_cafile
        self.openssl_capath_env = openssl_capath_env
        self.openssl_capath = openssl_capath

    def __repr__(self):
        return ("DefaultVerifyPaths(cafile=%r, capath=%r, openssl_cafile_env=%r, openssl_cafile=%r, "
                "openssl_capath_env=%r, openssl_capath=%r)" % (self.cafile, self.capath, self.openssl_cafile_env,
                                                                self.openssl_cafile, self.openssl_capath_env, self.openssl_capath))


def get_default_verify_paths():
    """Return paths to default cafile and capath."""
    openssl_cafile = "/etc/ssl/cert.pem"
    openssl_capath = "/etc/ssl/certs"
    cafile = os.environ.get("SSL_CERT_FILE", openssl_cafile)
    capath = os.environ.get("SSL_CERT_DIR", openssl_capath)
    return DefaultVerifyPaths(cafile if os.path.isfile(cafile) else None,
                              capath if os.path.isdir(capath) else None,
                              "SSL_CERT_FILE", openssl_cafile, "SSL_CERT_DIR", openssl_capath)


def RAND_bytes(n):
    return os.urandom(n)


def RAND_status():
    return True


def RAND_add(string, entropy):
    pass


class MemoryBIO:
    def __init__(self):
        raise NotImplementedError("py2go's ssl has no memory BIO")


class SSLObject:
    def __init__(self, *args, **kwargs):
        raise TypeError("SSLObject does not have a public constructor. Instances are returned by SSLContext.wrap_bio().")
