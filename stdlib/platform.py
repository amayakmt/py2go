"""platform: access to the underlying platform's identifying data."""
import os
import sys

__version__ = "1.0.8"


class uname_result:
    def __init__(self, system, node, release, version, machine, processor=""):
        self.system = system
        self.node = node
        self.release = release
        self.version = version
        self.machine = machine
        self.processor = processor

    def _t(self):
        return (self.system, self.node, self.release, self.version, self.machine)

    def __getitem__(self, i):
        return (self._t() + (self.processor,))[i]

    def __iter__(self):
        return iter(self._t() + (self.processor,))

    def __len__(self):
        return 6

    def __eq__(self, other):
        return tuple(self) == tuple(other)

    def __repr__(self):
        return ("uname_result(system=%r, node=%r, release=%r, version=%r, machine=%r)" % self._t())


_uname_cache = None


def uname():
    u = os.uname()
    return uname_result(u.sysname, u.nodename, u.release, u.version, u.machine, "")


def system():
    return uname().system


def node():
    return uname().node


def release():
    return uname().release


def version():
    return uname().version


def machine():
    return uname().machine


def processor():
    return uname().processor


def architecture(executable=sys.executable, bits="", linkage=""):
    return ("64bit", "ELF")


def python_implementation():
    return "CPython"


def python_version():
    return "3.13.0"


def python_version_tuple():
    return ("3", "13", "0")


def python_branch():
    return ""


def python_revision():
    return ""


def python_build():
    return ("main", "")


def python_compiler():
    return "Go"


def platform(aliased=False, terse=False):
    u = uname()
    if terse:
        return u.system + "-" + u.release
    return "%s-%s-%s" % (u.system, u.release, u.machine)


def libc_ver(executable=None, lib="", version="", chunksize=16384):
    return ("glibc", "2.0")


def mac_ver(release="", versioninfo=("", "", ""), machine=""):
    return (release, versioninfo, machine)


def win32_ver(release="", version="", csd="", ptype=""):
    return (release, version, csd, ptype)


def java_ver(release="", vendor="", vminfo=("", "", ""), osinfo=("", "", "")):
    return (release, vendor, vminfo, osinfo)


def freedesktop_os_release():
    info = {}
    try:
        with open("/etc/os-release", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                k, v = line.split("=", 1)
                if len(v) >= 2 and v[0] == v[-1] and v[0] in "\"'":
                    v = v[1:-1]
                info[k] = v
    except OSError:
        raise OSError(2, "No such file or directory: '/etc/os-release'")
    return info
