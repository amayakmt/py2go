"""resource: resource usage information."""
RLIMIT_CPU = 0
RLIMIT_FSIZE = 1
RLIMIT_DATA = 2
RLIMIT_STACK = 3
RLIMIT_CORE = 4
RLIMIT_RSS = 5
RLIMIT_NPROC = 6
RLIMIT_NOFILE = 7
RLIMIT_MEMLOCK = 8
RLIMIT_AS = 9
RLIM_INFINITY = -1
RUSAGE_SELF = 0
RUSAGE_CHILDREN = -1
RUSAGE_THREAD = 1

error = OSError


class struct_rusage:
    def __init__(self, t):
        self._t = tuple(t)
        self.ru_utime = t[0]
        self.ru_stime = t[1]
        self.ru_maxrss = t[2]
        self.ru_ixrss = t[3]
        self.ru_idrss = t[4]
        self.ru_isrss = t[5]
        self.ru_minflt = t[6]
        self.ru_majflt = t[7]
        self.ru_nswap = t[8]
        self.ru_inblock = t[9]
        self.ru_oublock = t[10]
        self.ru_msgsnd = t[11]
        self.ru_msgrcv = t[12]
        self.ru_nsignals = t[13]
        self.ru_nvcsw = t[14]
        self.ru_nivcsw = t[15]

    def __getitem__(self, i):
        return self._t[i]

    def __iter__(self):
        return iter(self._t)

    def __len__(self):
        return 16


def getrusage(who):
    return struct_rusage(_native("resource.getrusage", who))


def getrlimit(resource):
    return _native("resource.getrlimit", resource)


def setrlimit(resource, limits):
    _native("resource.setrlimit", resource, limits[0], limits[1])


def getpagesize():
    return 4096
