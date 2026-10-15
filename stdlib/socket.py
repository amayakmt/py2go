"""socket: low-level networking interface (BSD sockets on raw descriptors)."""
import io
import os
from builtins import socket_pm_gaierror as gaierror
from builtins import socket_pm_herror as herror
from builtins import TimeoutError as timeout

from builtins import OSError as error

AF_ALG = 38
AF_APPLETALK = 5
AF_ASH = 18
AF_ATMPVC = 8
AF_ATMSVC = 20
AF_AX25 = 3
AF_BRIDGE = 7
AF_CAN = 29
AF_ECONET = 19
AF_INET = 2
AF_INET6 = 10
AF_IPX = 4
AF_IRDA = 23
AF_KEY = 15
AF_LLC = 26
AF_NETBEUI = 13
AF_NETLINK = 16
AF_NETROM = 6
AF_PACKET = 17
AF_PPPOX = 24
AF_RDS = 21
AF_ROSE = 11
AF_ROUTE = 16
AF_SECURITY = 14
AF_SNA = 22
AF_TIPC = 30
AF_UNIX = 1
AF_UNSPEC = 0
AF_VSOCK = 40
AF_WANPIPE = 25
AF_X25 = 9
AI_ADDRCONFIG = 32
AI_ALL = 16
AI_CANONNAME = 2
AI_NUMERICHOST = 4
AI_NUMERICSERV = 1024
AI_PASSIVE = 1
AI_V4MAPPED = 8
ALG_OP_DECRYPT = 0
ALG_OP_ENCRYPT = 1
ALG_OP_SIGN = 2
ALG_OP_VERIFY = 3
ALG_SET_AEAD_ASSOCLEN = 4
ALG_SET_AEAD_AUTHSIZE = 5
ALG_SET_IV = 2
ALG_SET_KEY = 1
ALG_SET_OP = 3
ALG_SET_PUBKEY = 6
CAN_BCM = 2
CAN_BCM_CAN_FD_FRAME = 2048
CAN_BCM_RX_ANNOUNCE_RESUME = 256
CAN_BCM_RX_CHANGED = 12
CAN_BCM_RX_CHECK_DLC = 64
CAN_BCM_RX_DELETE = 6
CAN_BCM_RX_FILTER_ID = 32
CAN_BCM_RX_NO_AUTOTIMER = 128
CAN_BCM_RX_READ = 7
CAN_BCM_RX_RTR_FRAME = 1024
CAN_BCM_RX_SETUP = 5
CAN_BCM_RX_STATUS = 10
CAN_BCM_RX_TIMEOUT = 11
CAN_BCM_SETTIMER = 1
CAN_BCM_STARTTIMER = 2
CAN_BCM_TX_ANNOUNCE = 8
CAN_BCM_TX_COUNTEVT = 4
CAN_BCM_TX_CP_CAN_ID = 16
CAN_BCM_TX_DELETE = 2
CAN_BCM_TX_EXPIRED = 9
CAN_BCM_TX_READ = 3
CAN_BCM_TX_RESET_MULTI_IDX = 512
CAN_BCM_TX_SEND = 4
CAN_BCM_TX_SETUP = 1
CAN_BCM_TX_STATUS = 8
CAN_EFF_FLAG = 2147483648
CAN_EFF_MASK = 536870911
CAN_ERR_FLAG = 536870912
CAN_ERR_MASK = 536870911
CAN_ISOTP = 6
CAN_J1939 = 7
CAN_RAW = 1
CAN_RAW_ERR_FILTER = 2
CAN_RAW_FD_FRAMES = 5
CAN_RAW_FILTER = 1
CAN_RAW_JOIN_FILTERS = 6
CAN_RAW_LOOPBACK = 3
CAN_RAW_RECV_OWN_MSGS = 4
CAN_RTR_FLAG = 1073741824
CAN_SFF_MASK = 2047
EAGAIN = 11
EAI_ADDRFAMILY = -9
EAI_AGAIN = -3
EAI_BADFLAGS = -1
EAI_FAIL = -4
EAI_FAMILY = -6
EAI_MEMORY = -10
EAI_NODATA = -5
EAI_NONAME = -2
EAI_OVERFLOW = -12
EAI_SERVICE = -8
EAI_SOCKTYPE = -7
EAI_SYSTEM = -11
EBADF = 9
ETHERTYPE_ARP = 2054
ETHERTYPE_IP = 2048
ETHERTYPE_IPV6 = 34525
ETHERTYPE_VLAN = 33024
ETH_P_ALL = 3
EWOULDBLOCK = 11
INADDR_ALLHOSTS_GROUP = 3758096385
INADDR_ANY = 0
INADDR_BROADCAST = 4294967295
INADDR_LOOPBACK = 2130706433
INADDR_MAX_LOCAL_GROUP = 3758096639
INADDR_NONE = 4294967295
INADDR_UNSPEC_GROUP = 3758096384
IOCTL_VM_SOCKETS_GET_LOCAL_CID = 1977
IPPORT_RESERVED = 1024
IPPORT_USERRESERVED = 5000
IPPROTO_AH = 51
IPPROTO_DSTOPTS = 60
IPPROTO_EGP = 8
IPPROTO_ESP = 50
IPPROTO_FRAGMENT = 44
IPPROTO_GRE = 47
IPPROTO_HOPOPTS = 0
IPPROTO_ICMP = 1
IPPROTO_ICMPV6 = 58
IPPROTO_IDP = 22
IPPROTO_IGMP = 2
IPPROTO_IP = 0
IPPROTO_IPIP = 4
IPPROTO_IPV6 = 41
IPPROTO_NONE = 59
IPPROTO_PIM = 103
IPPROTO_PUP = 12
IPPROTO_RAW = 255
IPPROTO_ROUTING = 43
IPPROTO_RSVP = 46
IPPROTO_SCTP = 132
IPPROTO_TCP = 6
IPPROTO_TP = 29
IPPROTO_UDP = 17
IPPROTO_UDPLITE = 136
IPV6_CHECKSUM = 7
IPV6_DSTOPTS = 59
IPV6_HOPLIMIT = 52
IPV6_HOPOPTS = 54
IPV6_JOIN_GROUP = 20
IPV6_LEAVE_GROUP = 21
IPV6_MULTICAST_HOPS = 18
IPV6_MULTICAST_IF = 17
IPV6_MULTICAST_LOOP = 19
IPV6_NEXTHOP = 9
IPV6_PKTINFO = 50
IPV6_RECVDSTOPTS = 58
IPV6_RECVHOPLIMIT = 51
IPV6_RECVHOPOPTS = 53
IPV6_RECVPKTINFO = 49
IPV6_RECVRTHDR = 56
IPV6_RECVTCLASS = 66
IPV6_RTHDR = 57
IPV6_RTHDRDSTOPTS = 55
IPV6_RTHDR_TYPE_0 = 0
IPV6_TCLASS = 67
IPV6_UNICAST_HOPS = 16
IPV6_V6ONLY = 26
IP_ADD_MEMBERSHIP = 35
IP_ADD_SOURCE_MEMBERSHIP = 39
IP_BLOCK_SOURCE = 38
IP_DEFAULT_MULTICAST_LOOP = 1
IP_DEFAULT_MULTICAST_TTL = 1
IP_DROP_MEMBERSHIP = 36
IP_DROP_SOURCE_MEMBERSHIP = 40
IP_HDRINCL = 3
IP_MAX_MEMBERSHIPS = 20
IP_MULTICAST_IF = 32
IP_MULTICAST_LOOP = 34
IP_MULTICAST_TTL = 33
IP_OPTIONS = 4
IP_PKTINFO = 8
IP_RECVOPTS = 6
IP_RECVRETOPTS = 7
IP_RECVTOS = 13
IP_RETOPTS = 7
IP_TOS = 1
IP_TRANSPARENT = 19
IP_TTL = 2
IP_UNBLOCK_SOURCE = 37
J1939_EE_INFO_NONE = 0
J1939_EE_INFO_TX_ABORT = 1
J1939_FILTER_MAX = 512
J1939_IDLE_ADDR = 254
J1939_MAX_UNICAST_ADDR = 253
J1939_NLA_BYTES_ACKED = 1
J1939_NLA_PAD = 0
J1939_NO_ADDR = 255
J1939_NO_NAME = 0
J1939_NO_PGN = 262144
J1939_PGN_ADDRESS_CLAIMED = 60928
J1939_PGN_ADDRESS_COMMANDED = 65240
J1939_PGN_MAX = 262143
J1939_PGN_PDU1_MAX = 261888
J1939_PGN_REQUEST = 59904
MSG_CMSG_CLOEXEC = 1073741824
MSG_CONFIRM = 2048
MSG_CTRUNC = 8
MSG_DONTROUTE = 4
MSG_DONTWAIT = 64
MSG_EOR = 128
MSG_ERRQUEUE = 8192
MSG_FASTOPEN = 536870912
MSG_MORE = 32768
MSG_NOSIGNAL = 16384
MSG_OOB = 1
MSG_PEEK = 2
MSG_TRUNC = 32
MSG_WAITALL = 256
NETLINK_CRYPTO = 21
NETLINK_DNRTMSG = 14
NETLINK_FIREWALL = 3
NETLINK_IP6_FW = 13
NETLINK_NFLOG = 5
NETLINK_ROUTE = 0
NETLINK_USERSOCK = 2
NETLINK_XFRM = 6
NI_DGRAM = 16
NI_IDN = 32
NI_MAXHOST = 1025
NI_MAXSERV = 32
NI_NAMEREQD = 8
NI_NOFQDN = 4
NI_NUMERICHOST = 1
NI_NUMERICSERV = 2
PACKET_BROADCAST = 1
PACKET_FASTROUTE = 6
PACKET_HOST = 0
PACKET_LOOPBACK = 5
PACKET_MULTICAST = 2
PACKET_OTHERHOST = 3
PACKET_OUTGOING = 4
PF_CAN = 29
PF_PACKET = 17
PF_RDS = 21
SCM_CREDENTIALS = 2
SCM_J1939_DEST_ADDR = 1
SCM_J1939_DEST_NAME = 2
SCM_J1939_ERRQUEUE = 4
SCM_J1939_PRIO = 3
SCM_RIGHTS = 1
SHUT_RD = 0
SHUT_RDWR = 2
SHUT_WR = 1
SOCK_CLOEXEC = 524288
SOCK_DGRAM = 2
SOCK_NONBLOCK = 2048
SOCK_RAW = 3
SOCK_RDM = 4
SOCK_SEQPACKET = 5
SOCK_STREAM = 1
SOL_ALG = 279
SOL_CAN_BASE = 100
SOL_CAN_RAW = 101
SOL_IP = 0
SOL_SOCKET = 1
SOL_TCP = 6
SOL_TIPC = 271
SOL_UDP = 17
SOMAXCONN = 128
SO_ACCEPTCONN = 30
SO_BINDTODEVICE = 25
SO_BINDTOIFINDEX = 62
SO_BROADCAST = 6
SO_DEBUG = 1
SO_DOMAIN = 39
SO_DONTROUTE = 5
SO_ERROR = 4
SO_INCOMING_CPU = 49
SO_J1939_ERRQUEUE = 4
SO_J1939_FILTER = 1
SO_J1939_PROMISC = 2
SO_J1939_SEND_PRIO = 3
SO_KEEPALIVE = 9
SO_LINGER = 13
SO_MARK = 36
SO_OOBINLINE = 10
SO_PASSCRED = 16
SO_PASSSEC = 34
SO_PEERCRED = 17
SO_PEERSEC = 31
SO_PRIORITY = 12
SO_PROTOCOL = 38
SO_RCVBUF = 8
SO_RCVLOWAT = 18
SO_RCVTIMEO = 20
SO_REUSEADDR = 2
SO_REUSEPORT = 15
SO_SNDBUF = 7
SO_SNDLOWAT = 19
SO_SNDTIMEO = 21
SO_TYPE = 3
SO_VM_SOCKETS_BUFFER_MAX_SIZE = 2
SO_VM_SOCKETS_BUFFER_MIN_SIZE = 1
SO_VM_SOCKETS_BUFFER_SIZE = 0
TCP_CONGESTION = 13
TCP_CORK = 3
TCP_DEFER_ACCEPT = 9
TCP_FASTOPEN = 23
TCP_INFO = 11
TCP_KEEPCNT = 6
TCP_KEEPIDLE = 4
TCP_KEEPINTVL = 5
TCP_LINGER2 = 8
TCP_MAXSEG = 2
TCP_MD5SIG = 14
TCP_NODELAY = 1
TCP_QUEUE_SEQ = 21
TCP_QUICKACK = 12
TCP_REPAIR = 19
TCP_REPAIR_OPTIONS = 22
TCP_REPAIR_QUEUE = 20
TCP_SYNCNT = 7
TCP_THIN_DUPACK = 17
TCP_THIN_LINEAR_TIMEOUTS = 16
TCP_TIMESTAMP = 24
TCP_USER_TIMEOUT = 18
TCP_WINDOW_CLAMP = 10
TIPC_ADDR_ID = 3
TIPC_ADDR_NAME = 2
TIPC_ADDR_NAMESEQ = 1
TIPC_CFG_SRV = 0
TIPC_CLUSTER_SCOPE = 2
TIPC_CONN_TIMEOUT = 130
TIPC_CRITICAL_IMPORTANCE = 3
TIPC_DEST_DROPPABLE = 129
TIPC_HIGH_IMPORTANCE = 2
TIPC_IMPORTANCE = 127
TIPC_LOW_IMPORTANCE = 0
TIPC_MEDIUM_IMPORTANCE = 1
TIPC_NODE_SCOPE = 3
TIPC_PUBLISHED = 1
TIPC_SRC_DROPPABLE = 128
TIPC_SUBSCR_TIMEOUT = 3
TIPC_SUB_CANCEL = 4
TIPC_SUB_PORTS = 1
TIPC_SUB_SERVICE = 2
TIPC_TOP_SRV = 1
TIPC_WAIT_FOREVER = -1
TIPC_WITHDRAWN = 2
TIPC_ZONE_SCOPE = 1
UDPLITE_RECV_CSCOV = 11
UDPLITE_SEND_CSCOV = 10
VMADDR_CID_ANY = 4294967295
VMADDR_CID_HOST = 2
VMADDR_PORT_ANY = 4294967295
VM_SOCKETS_INVALID_VERSION = 4294967295

has_ipv6 = True
EBADF = 9
EAGAIN = 11
EWOULDBLOCK = 11

_GLOBAL_DEFAULT_TIMEOUT = -1.0
_defaulttimeout = None


def getdefaulttimeout():
    return _defaulttimeout


def setdefaulttimeout(timeout):
    global _defaulttimeout
    if timeout is None:
        _defaulttimeout = None
        return
    t = float(timeout)
    if t < 0:
        raise ValueError("Timeout value out of range")
    _defaulttimeout = t


def _obj(x: object) -> object:
    return x


class socket:
    def __init__(self, family=-1, type=-1, proto=-1, fileno=None):
        self._timeout = _obj(_defaulttimeout)
        self._closed = False
        self._io_refs = 0
        self._sock_closed = False
        if fileno is None:
            if family == -1:
                family = AF_INET
            if type == -1:
                type = SOCK_STREAM
            if proto == -1:
                proto = 0
            self._fd = _native("_socket.socket", family, type & ~(SOCK_NONBLOCK | SOCK_CLOEXEC), proto)
            if type & SOCK_NONBLOCK:
                self._timeout = _obj(0.0)
        else:
            if not isinstance(fileno, int):
                raise TypeError("an integer is required")
            if fileno < 0:
                raise ValueError("negative file descriptor")
            self._fd = fileno
            if family == -1:
                family = _native("_socket.getsockopt", fileno, SOL_SOCKET, 39, None)
            if type == -1:
                type = _native("_socket.getsockopt", fileno, SOL_SOCKET, SO_TYPE, None)
            if proto == -1:
                proto = _native("_socket.getsockopt", fileno, SOL_SOCKET, 38, None)
        self.family = family
        self.type = type & 0xF
        self.proto = proto
        self._apply_timeout()

    def _apply_timeout(self):
        t = self._timeout
        if t is None:
            _native("_socket.set_nonblock", self._fd, False)
        else:
            _native("_socket.set_nonblock", self._fd, True)

    def _tmo(self):
        t = self._timeout
        if t is None:
            return None
        return t

    def _check(self):
        if self._sock_closed:
            raise OSError(9, "Bad file descriptor")

    def __repr__(self):
        closed = "[closed] " if self._closed else ""
        s = "<socket.socket %sfd=%i, family=%s, type=%s, proto=%i" % (closed, self.fileno(), self.family, self.type, self.proto)
        if not self._closed:
            try:
                laddr = self.getsockname()
                if laddr:
                    s += ", laddr=%s" % str(laddr)
            except OSError:
                pass
            try:
                raddr = self.getpeername()
                if raddr:
                    s += ", raddr=%s" % str(raddr)
            except OSError:
                pass
        s += ">"
        return s

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        if not self._closed:
            self.close()

    def fileno(self):
        if self._closed:
            return -1
        return self._fd

    def detach(self):
        self._closed = True
        fd = self._fd
        self._fd = -1
        return fd

    def dup(self):
        fd = os.dup(self._fd)
        s = socket(self.family, self.type, self.proto, fd)
        s._timeout = _obj(self._timeout)
        s._apply_timeout()
        return s

    def close(self):
        self._closed = True
        if self._io_refs <= 0:
            self._real_close()

    def _real_close(self):
        if not self._sock_closed:
            self._sock_closed = True
            fd = self._fd
            self._fd = -1
            if fd >= 0:
                try:
                    _native("_socket.close", fd)
                except OSError as e:
                    if e.errno != 104:
                        raise

    def _decref_socketios(self):
        if self._io_refs > 0:
            self._io_refs -= 1
        if self._closed:
            self._real_close()

    def settimeout(self, value):
        if value is None:
            self._timeout = _obj(None)
        else:
            if not isinstance(value, (int, float)):
                raise TypeError("Timeout value must be a float or None")
            if value < 0:
                raise ValueError("Timeout value out of range")
            self._timeout = _obj(float(value))
        self._check()
        self._apply_timeout()

    def gettimeout(self):
        return self._timeout

    def setblocking(self, flag):
        if flag:
            self._timeout = _obj(None)
        else:
            self._timeout = _obj(0.0)
        self._check()
        self._apply_timeout()

    def getblocking(self):
        if self._timeout is None:
            return True
        return self._timeout != 0.0

    def bind(self, address):
        self._check()
        _native("_socket.bind", self._fd, self.family, address)

    def listen(self, backlog=None):
        self._check()
        _native("_socket.listen", self._fd, backlog)

    def accept(self):
        self._check()
        fd, addr = _native("_socket.accept", self._fd, self._timeout)
        sock = socket(self.family, self.type, self.proto, fd)
        if getdefaulttimeout() is None and self._timeout:
            sock.setblocking(True)
        return sock, addr

    def connect(self, address):
        self._check()
        _native("_socket.connect", self._fd, self.family, address, self._timeout)

    def connect_ex(self, address):
        try:
            self.connect(address)
        except OSError as e:
            if e.errno is None:
                raise
            return e.errno
        return 0

    def getsockname(self):
        self._check()
        return _native("_socket.getsockname", self._fd)

    def getpeername(self):
        self._check()
        return _native("_socket.getpeername", self._fd)

    def recv(self, bufsize, flags=0):
        self._check()
        return _native("_socket.recv", self._fd, bufsize, flags, self._timeout)

    def recv_into(self, buffer, nbytes=0, flags=0):
        if nbytes < 0:
            raise ValueError("negative buffersize in recv_into")
        if nbytes == 0:
            nbytes = len(buffer)
        data = self.recv(nbytes, flags)
        n = len(data)
        buffer[:n] = data
        return n

    def recvfrom(self, bufsize, flags=0):
        self._check()
        return _native("_socket.recvfrom", self._fd, bufsize, flags, self._timeout)

    def recvfrom_into(self, buffer, nbytes=0, flags=0):
        if nbytes == 0:
            nbytes = len(buffer)
        data, addr = self.recvfrom(nbytes, flags)
        n = len(data)
        buffer[:n] = data
        return n, addr

    def send(self, data, flags=0):
        self._check()
        return _native("_socket.send", self._fd, data, flags, self._timeout)

    def sendall(self, data, flags=0):
        self._check()
        data = bytes(data)
        total = len(data)
        sent = 0
        while sent < total:
            n = _native("_socket.send", self._fd, data[sent:], flags, self._timeout)
            sent += n
        return None

    def sendto(self, data, flags_or_addr, addr=None):
        self._check()
        if addr is None:
            flags = 0
            addr = flags_or_addr
        else:
            flags = flags_or_addr
        return _native("_socket.sendto", self._fd, data, flags, self.family, addr, self._timeout)

    def sendmsg(self, buffers, ancdata=(), flags=0, address=None):
        self._check()
        return _native("_socket.sendmsg", self._fd, buffers, ancdata, flags, self.family, address, self._timeout)

    def recvmsg(self, bufsize, ancbufsize=0, flags=0):
        self._check()
        return _native("_socket.recvmsg", self._fd, bufsize, ancbufsize, flags, self._timeout)

    def shutdown(self, how):
        self._check()
        _native("_socket.shutdown", self._fd, how)

    def setsockopt(self, level, optname, value, optlen=None):
        self._check()
        _native("_socket.setsockopt", self._fd, level, optname, value)

    def getsockopt(self, level, optname, buflen=None):
        self._check()
        return _native("_socket.getsockopt", self._fd, level, optname, buflen)

    def get_inheritable(self):
        return False

    def set_inheritable(self, inheritable):
        pass

    def makefile(self, mode="r", buffering=None, *, encoding=None, errors=None, newline=None):
        for c in mode:
            if c not in ("r", "w", "b"):
                raise ValueError("invalid mode %r (only r, w, b allowed)" % (mode,))
        writing = "w" in mode
        reading = "r" in mode or not writing
        binary = "b" in mode
        rawmode = ""
        if reading:
            rawmode += "r"
        if writing:
            rawmode += "w"
        raw = SocketIO(self, rawmode)
        self._io_refs += 1
        if buffering is None:
            buffering = -1
        if buffering < 0:
            buffering = io.DEFAULT_BUFFER_SIZE
        if buffering == 0:
            if not binary:
                raise ValueError("unbuffered streams must be binary")
            return raw
        if reading and writing:
            buffer = io.BufferedRWPair(raw, raw, buffering)
        elif reading:
            buffer = io.BufferedReader(raw, buffering)
        else:
            buffer = io.BufferedWriter(raw, buffering)
        if binary:
            return buffer
        encoding = io.text_encoding(encoding)
        text = io.TextIOWrapper(buffer, encoding, errors, newline)
        return text


class SocketIO(io.RawIOBase):
    def __init__(self, sock, mode):
        io.RawIOBase.__init__(self)
        self._sock = sock
        self._mode = mode
        self._reading = "r" in mode
        self._writing = "w" in mode
        self._timeout_occurred = False
        self._rclosed = False

    def readinto(self, b):
        if self._rclosed:
            raise ValueError("I/O operation on closed file.")
        if not self._reading:
            raise OSError("File not open for reading")
        if self._timeout_occurred:
            raise OSError("cannot read from timed out object")
        try:
            data = self._sock.recv(len(b))
        except timeout:
            self._timeout_occurred = True
            raise
        n = len(data)
        b[:n] = data
        return n

    def read(self, size=-1):
        if self._rclosed:
            raise ValueError("I/O operation on closed file.")
        if not self._reading:
            raise OSError("File not open for reading")
        if size is None or size < 0:
            parts = []
            while True:
                chunk = self._sock.recv(8192)
                if not chunk:
                    break
                parts.append(chunk)
            return b"".join(parts)
        if self._timeout_occurred:
            raise OSError("cannot read from timed out object")
        try:
            return self._sock.recv(size)
        except timeout:
            self._timeout_occurred = True
            raise

    def write(self, b):
        if self._rclosed:
            raise ValueError("I/O operation on closed file.")
        if not self._writing:
            raise OSError("File not open for writing")
        data = bytes(b)
        return self._sock.send(data)

    def readable(self):
        return self._reading and not self._rclosed

    def writable(self):
        return self._writing and not self._rclosed

    def seekable(self):
        return False

    def fileno(self):
        return self._sock.fileno()

    @property
    def name(self):
        if not self._rclosed:
            return self.fileno()
        return -1

    @property
    def mode(self):
        return self._mode

    @property
    def closed(self):
        return self._rclosed

    def flush(self):
        pass

    def close(self):
        if self._rclosed:
            return
        self._rclosed = True
        self._sock._decref_socketios()


def socketpair(family=None, type=SOCK_STREAM, proto=0):
    if family is None:
        family = AF_UNIX
    a, b = _native("_socket.socketpair", family, type, proto)
    return socket(family, type, proto, a), socket(family, type, proto, b)


def fromfd(fd, family, type, proto=0):
    return socket(family, type, proto, os.dup(fd))


def getaddrinfo(host, port, family=0, type=0, proto=0, flags=0):
    return _native("_socket.getaddrinfo", host, port, family, type, proto, flags)


def gethostname():
    return _native("_socket.gethostname")


def gethostbyname(hostname):
    return _native("_socket.gethostbyname_ex", hostname)[2][0]


def gethostbyname_ex(hostname):
    return _native("_socket.gethostbyname_ex", hostname)


def gethostbyaddr(ip_address):
    return _native("_socket.gethostbyaddr", ip_address)


def getfqdn(name=""):
    name = name.strip()
    if not name or name in ("0.0.0.0", "::"):
        name = gethostname()
    try:
        hostname, aliases, ipaddrs = gethostbyaddr(name)
    except (OSError, herror):
        pass
    else:
        aliases.insert(0, hostname)
        for name in aliases:
            if "." in name:
                break
        else:
            name = hostname
    return name


def getservbyname(servicename, protocolname=None):
    return _native("_socket.getservbyname", servicename, protocolname)


def getservbyport(port, protocolname=None):
    return _native("_socket.getservbyport", port, protocolname)


def getprotobyname(protocolname):
    return _native("_socket.getprotobyname", protocolname)


def getnameinfo(sockaddr, flags):
    host = sockaddr[0]
    port = sockaddr[1]
    if not flags & NI_NUMERICHOST:
        try:
            host = gethostbyaddr(host)[0]
        except (OSError, herror):
            pass
    if not flags & NI_NUMERICSERV:
        try:
            for name, n in (("http", 80), ("https", 443), ("ftp", 21), ("ssh", 22), ("smtp", 25)):
                if n == port:
                    return host, name
        except OSError:
            pass
    return host, str(port)


def inet_aton(ip_string):
    return _native("_socket.inet_aton", ip_string)


def inet_ntoa(packed_ip):
    packed_ip = bytes(packed_ip)
    if len(packed_ip) != 4:
        raise OSError("packed IP wrong length for inet_ntoa")
    return "%d.%d.%d.%d" % (packed_ip[0], packed_ip[1], packed_ip[2], packed_ip[3])


def inet_pton(address_family, ip_string):
    return _native("_socket.inet_pton", address_family, ip_string)


def inet_ntop(address_family, packed_ip):
    return _native("_socket.inet_ntop", address_family, packed_ip)


def htons(x):
    if x < 0 or x > 65535:
        raise OverflowError("htons: Python int too large to convert to C 16-bit unsigned integer")
    return ((x & 0xFF) << 8) | (x >> 8)


def ntohs(x):
    return htons(x)


def htonl(x):
    if x < 0 or x > 4294967295:
        raise OverflowError("htonl: Python int too large to convert to C 32-bit unsigned integer")
    return ((x & 0xFF) << 24) | ((x & 0xFF00) << 8) | ((x >> 8) & 0xFF00) | (x >> 24)


def ntohl(x):
    return htonl(x)


def if_nameindex():
    return _native("_socket.if_nameindex")


def if_nametoindex(name):
    return _native("_socket.if_nametoindex", name)


def if_indextoname(index):
    return _native("_socket.if_indextoname", index)


def CMSG_LEN(length):
    return 16 + length


def CMSG_SPACE(length):
    return 16 + ((length + 7) // 8) * 8


def close(fd):
    _native("_socket.close", fd)


def has_dualstack_ipv6():
    return True


def create_connection(address, timeout=_GLOBAL_DEFAULT_TIMEOUT, source_address=None, *, all_errors=False):
    host, port = address
    exceptions = []
    for res in getaddrinfo(host, port, 0, SOCK_STREAM):
        af, socktype, proto, canonname, sa = res
        sock = None
        try:
            sock = socket(af, socktype, proto)
            if timeout is not _GLOBAL_DEFAULT_TIMEOUT and timeout != _GLOBAL_DEFAULT_TIMEOUT:
                sock.settimeout(timeout)
            if source_address:
                sock.bind(source_address)
            sock.connect(sa)
            exceptions.clear()
            return sock
        except error as exc:
            if not all_errors:
                exceptions.clear()
            exceptions.append(exc)
            if sock is not None:
                sock.close()
    if len(exceptions):
        try:
            if not all_errors:
                raise exceptions[0]
            raise ExceptionGroup("create_connection failed", exceptions)
        finally:
            exceptions.clear()
    else:
        raise error("getaddrinfo returns an empty list")


def create_server(address, *, family=AF_INET, backlog=None, reuse_port=False, dualstack_ipv6=False):
    if reuse_port and not hasattr_reuseport():
        raise ValueError("SO_REUSEPORT not supported on this platform")
    if dualstack_ipv6:
        if not has_dualstack_ipv6():
            raise ValueError("dualstack_ipv6 not supported on this platform")
        if family != AF_INET6:
            raise ValueError("dualstack_ipv6 requires AF_INET6 family")
    sock = socket(family, SOCK_STREAM)
    try:
        if os.name not in ("nt", "cygwin"):
            try:
                sock.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)
            except error:
                pass
        if reuse_port:
            sock.setsockopt(SOL_SOCKET, SO_REUSEPORT, 1)
        if has_ipv6 and family == AF_INET6:
            if dualstack_ipv6:
                sock.setsockopt(IPPROTO_IPV6, IPV6_V6ONLY, 0)
            else:
                sock.setsockopt(IPPROTO_IPV6, IPV6_V6ONLY, 1)
        try:
            sock.bind(address)
        except error as err:
            msg = "%s (while attempting to bind on address %r)" % (err.strerror, address)
            raise error(err.errno, msg) from None
        if backlog is None:
            sock.listen()
        else:
            sock.listen(backlog)
        return sock
    except error:
        sock.close()
        raise


def hasattr_reuseport():
    return True
