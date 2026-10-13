"""random: pseudo-random numbers. The generator is CPython's Mersenne
Twister, seeded and drawn from the same way, so a seeded program gets the
numbers CPython gives it."""
import math
import time
from hashlib import sha512 as _sha512
from itertools import accumulate as _accumulate
from bisect import bisect as _bisect

NV_MAGICCONST = 4 * math.exp(-0.5) / math.sqrt(2.0)
TWOPI = 2.0 * math.pi
LOG4 = math.log(4.0)
SG_MAGICCONST = 1.0 + math.log(4.5)
BPF = 53
RECIP_BPF = 2 ** -BPF


def _index(x):
    if isinstance(x, int):
        return x
    if isinstance(x, float):
        raise TypeError("'float' object cannot be interpreted as an integer")
    return x.__index__()


class Random:
    VERSION = 3

    def __init__(self, x=None):
        self._mt = _native("random.mt_new")
        self.gauss_next = None
        self.seed(x)

    def seed(self, a=None, version=2):
        value = 0
        if a is None:
            value = time.time_ns() ^ (id(self) << 20)
        elif version == 2 and isinstance(a, (str, bytes, bytearray)):
            data = a.encode() if isinstance(a, str) else bytes(a)
            value = int.from_bytes(data + _sha512(data).digest(), "big")
        elif isinstance(a, float):
            value = hash(a)
        elif isinstance(a, int):
            value = a
        else:
            raise TypeError("The only supported seed types are:\nNone, int, float, str, bytes, and bytearray.")
        _native("random.mt_seed", self._mt, value)
        self.gauss_next = None

    def random(self):
        return _native("random.mt_random", self._mt)

    def getrandbits(self, k):
        if k < 0:
            raise ValueError("number of bits must be non-negative")
        return _native("random.mt_bits", self._mt, k)

    def randbytes(self, n):
        return self.getrandbits(n * 8).to_bytes(n, "little")

    def getstate(self):
        return (self.VERSION, _native("random.mt_getstate", self._mt), self.gauss_next)

    def setstate(self, state):
        version = state[0]
        if version != 3 and version != 2:
            raise ValueError("state with version " + str(version) + " passed to Random.setstate() of version 3")
        _native("random.mt_setstate", self._mt, state[1])
        self.gauss_next = state[2]

    def _randbelow(self, n):
        if not n:
            return 0
        k = n.bit_length()
        r = self.getrandbits(k)
        while r >= n:
            r = self.getrandbits(k)
        return r

    def randrange(self, start, stop=None, step=1):
        istart = _index(start)
        if stop is None:
            if step != 1:
                raise TypeError("Missing a non-None stop argument")
            if istart > 0:
                return self._randbelow(istart)
            raise ValueError("empty range for randrange()")
        istop = _index(stop)
        width = istop - istart
        istep = _index(step)
        if istep == 1:
            if width > 0:
                return istart + self._randbelow(width)
            raise ValueError("empty range in randrange(" + str(start) + ", " + str(stop) + ")")
        if istep > 0:
            n = (width + istep - 1) // istep
        elif istep < 0:
            n = (width + istep + 1) // istep
        else:
            raise ValueError("zero step for randrange()")
        if n <= 0:
            raise ValueError("empty range in randrange(" + str(start) + ", " + str(stop) + ", " + str(step) + ")")
        return istart + istep * self._randbelow(n)

    def randint(self, a, b):
        return self.randrange(a, b + 1)

    def choice(self, seq):
        if not len(seq):
            raise IndexError("Cannot choose from an empty sequence")
        return seq[self._randbelow(len(seq))]

    def shuffle(self, x):
        for i in reversed(range(1, len(x))):
            j = self._randbelow(i + 1)
            x[i], x[j] = x[j], x[i]

    def sample(self, population, k, *, counts=None):
        if isinstance(population, (set, frozenset, dict)):
            raise TypeError("Population must be a sequence.  For dicts or sets, use sorted(d).")
        n = len(population)
        if counts is not None:
            cum_counts = list(_accumulate(counts))
            if len(cum_counts) != n:
                raise ValueError("The number of counts does not match the population")
            total = cum_counts.pop()
            if not isinstance(total, int):
                raise TypeError("Counts must be integers")
            if total <= 0:
                raise ValueError("Total of counts must be greater than zero")
            selections = self.sample(range(total), k=k)
            return [population[_bisect(cum_counts, s)] for s in selections]
        if not 0 <= k <= n:
            raise ValueError("Sample larger than population or is negative")
        result = [None] * k
        setsize = 21
        if k > 5:
            setsize += 4 ** math.ceil(math.log(k * 3, 4))
        if n <= setsize:
            pool = list(population)
            for i in range(k):
                j = self._randbelow(n - i)
                result[i] = pool[j]
                pool[j] = pool[n - i - 1]
        else:
            selected = set()
            for i in range(k):
                j = self._randbelow(n)
                while j in selected:
                    j = self._randbelow(n)
                selected.add(j)
                result[i] = population[j]
        return result

    def choices(self, population, weights=None, *, cum_weights=None, k=1):
        n = len(population)
        if cum_weights is None:
            if weights is None:
                n += 0.0
                return [population[math.floor(self.random() * n)] for i in range(k)]
            cum_weights = list(_accumulate(weights))
        elif weights is not None:
            raise TypeError("Cannot specify both weights and cumulative weights")
        if len(cum_weights) != n:
            raise ValueError("The number of weights does not match the population")
        total = cum_weights[-1] + 0.0
        if total <= 0.0:
            raise ValueError("Total of weights must be greater than zero")
        if not math.isfinite(total):
            raise ValueError("Total of weights must be finite")
        hi = n - 1
        return [population[_bisect(cum_weights, self.random() * total, 0, hi)] for i in range(k)]

    def uniform(self, a, b):
        return a + (b - a) * self.random()

    def triangular(self, low=0.0, high=1.0, mode=None):
        u = self.random()
        try:
            c = 0.5 if mode is None else (mode - low) / (high - low)
        except ZeroDivisionError:
            return low
        if u > c:
            u = 1.0 - u
            c = 1.0 - c
            low, high = high, low
        return low + (high - low) * math.sqrt(u * c)

    def normalvariate(self, mu=0.0, sigma=1.0):
        while True:
            u1 = self.random()
            u2 = 1.0 - self.random()
            z = NV_MAGICCONST * (u1 - 0.5) / u2
            zz = z * z / 4.0
            if zz <= -math.log(u2):
                break
        return mu + z * sigma

    def gauss(self, mu=0.0, sigma=1.0):
        z = self.gauss_next
        self.gauss_next = None
        if z is None:
            x2pi = self.random() * TWOPI
            g2rad = math.sqrt(-2.0 * math.log(1.0 - self.random()))
            z = math.cos(x2pi) * g2rad
            self.gauss_next = math.sin(x2pi) * g2rad
        return mu + z * sigma

    def lognormvariate(self, mu, sigma):
        return math.exp(self.normalvariate(mu, sigma))

    def expovariate(self, lambd=1.0):
        return -math.log(1.0 - self.random()) / lambd

    def vonmisesvariate(self, mu, kappa):
        if kappa <= 1e-6:
            return TWOPI * self.random()
        s = 0.5 / kappa
        r = s + math.sqrt(1.0 + s * s)
        while True:
            u1 = self.random()
            z = math.cos(math.pi * u1)
            d = z / (r + z)
            u2 = self.random()
            if u2 < 1.0 - d * d or u2 <= (1.0 - d) * math.exp(d):
                break
        q = 1.0 / r
        f = (q + z) / (1.0 + q * z)
        u3 = self.random()
        if u3 > 0.5:
            theta = (mu + math.acos(f)) % TWOPI
        else:
            theta = (mu - math.acos(f)) % TWOPI
        return theta

    def gammavariate(self, alpha, beta):
        if alpha <= 0.0 or beta <= 0.0:
            raise ValueError("gammavariate: alpha and beta must be > 0.0")
        if alpha > 1.0:
            ainv = math.sqrt(2.0 * alpha - 1.0)
            bbb = alpha - LOG4
            ccc = alpha + ainv
            while True:
                u1 = self.random()
                if not 1e-7 < u1 < 0.9999999:
                    continue
                u2 = 1.0 - self.random()
                v = math.log(u1 / (1.0 - u1)) / ainv
                x = alpha * math.exp(v)
                z = u1 * u1 * u2
                r = bbb + ccc * v - x
                if r + SG_MAGICCONST - 4.5 * z >= 0.0 or r >= math.log(z):
                    return x * beta
        elif alpha == 1.0:
            return -math.log(1.0 - self.random()) * beta
        else:
            while True:
                u = self.random()
                b = (math.e + alpha) / math.e
                p = b * u
                if p <= 1.0:
                    x = p ** (1.0 / alpha)
                else:
                    x = -math.log((b - p) / alpha)
                u1 = self.random()
                if p > 1.0:
                    if u1 <= x ** (alpha - 1.0):
                        break
                elif u1 <= math.exp(-x):
                    break
            return x * beta

    def betavariate(self, alpha, beta):
        y = self.gammavariate(alpha, 1.0)
        if y:
            return y / (y + self.gammavariate(beta, 1.0))
        return 0.0

    def paretovariate(self, alpha):
        u = 1.0 - self.random()
        return u ** (-1.0 / alpha)

    def weibullvariate(self, alpha, beta):
        u = 1.0 - self.random()
        return alpha * (-math.log(u)) ** (1.0 / beta)

    def binomialvariate(self, n=1, p=0.5):
        if n < 0:
            raise ValueError("n must be non-negative")
        if p <= 0.0 or p >= 1.0:
            if p == 0.0:
                return 0
            if p == 1.0:
                return n
            raise ValueError("p must be in the range 0.0 <= p <= 1.0")
        total = 0
        for _ in range(n):
            if self.random() < p:
                total += 1
        return total


_inst = Random()
seed = _inst.seed
random = _inst.random
uniform = _inst.uniform
triangular = _inst.triangular
randint = _inst.randint
choice = _inst.choice
randrange = _inst.randrange
sample = _inst.sample
shuffle = _inst.shuffle
choices = _inst.choices
normalvariate = _inst.normalvariate
lognormvariate = _inst.lognormvariate
expovariate = _inst.expovariate
vonmisesvariate = _inst.vonmisesvariate
gammavariate = _inst.gammavariate
gauss = _inst.gauss
betavariate = _inst.betavariate
paretovariate = _inst.paretovariate
weibullvariate = _inst.weibullvariate
getstate = _inst.getstate
setstate = _inst.setstate
getrandbits = _inst.getrandbits
randbytes = _inst.randbytes
binomialvariate = _inst.binomialvariate
