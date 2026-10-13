"""statistics: mathematical statistics functions."""

import math
import sys
import random
from fractions import Fraction
from collections import Counter
from collections import namedtuple
from functools import reduce
from itertools import groupby
from bisect import bisect_left
from bisect import bisect_right

from builtins import statistics_pm_StatisticsError as StatisticsError

_sqrt_bit_width = 2 * 53 + 3


def _isfinite(x):
    try:
        return x.is_finite()
    except AttributeError:
        return math.isfinite(x)


def _coerce(T, S):
    if T is S:
        return T
    if S is int or S is bool:
        return T
    if T is int:
        return S
    if issubclass(S, T):
        return S
    if issubclass(T, S):
        return T
    if issubclass(T, int):
        return S
    if issubclass(S, int):
        return T
    if issubclass(T, Fraction) and issubclass(S, float):
        return S
    if issubclass(T, float) and issubclass(S, Fraction):
        return T
    raise TypeError("don't know how to coerce " + T.__name__ + " and " + S.__name__)


def _exact_ratio(x):
    if isinstance(x, float):
        if math.isnan(x) or math.isinf(x):
            return (x, None)
        return x.as_integer_ratio()
    if isinstance(x, int):
        return (x, 1)
    try:
        return x.as_integer_ratio()
    except AttributeError:
        pass
    except (OverflowError, ValueError):
        return (x, None)
    try:
        return (x.numerator, x.denominator)
    except AttributeError:
        msg = "can't convert type '" + type(x).__name__ + "' to numerator/denominator"
        raise TypeError(msg)


def _sum(data):
    count = 0
    types = set()
    partials = {}
    for v in data:
        types.add(type(v))
        pair = _exact_ratio(v)
        n, d = pair[0], pair[1]
        count += 1
        partials[d] = partials.get(d, 0) + n
    if None in partials:
        total = partials[None]
    else:
        total = Fraction(0)
        for d, n in partials.items():
            total = total + Fraction(n, d)
    T = int
    for t in types:
        T = _coerce(T, t)
    return (T, total, count)


def _convert(value, T):
    if type(value) is T:
        return value
    if issubclass(T, int) and value.denominator != 1:
        T = float
    return T(value)


def _fail_neg(values, errmsg="negative value"):
    out = []
    for x in values:
        if x < 0:
            raise StatisticsError(errmsg)
        out.append(x)
    return out


def _integer_sqrt_of_frac_rto(n, m):
    a = math.isqrt(n // m)
    return a + (a * a * m != n)


def _float_sqrt_of_frac(n, m):
    q = (n.bit_length() - m.bit_length() - _sqrt_bit_width) // 2
    if q >= 0:
        numerator = _integer_sqrt_of_frac_rto(n, m << 2 * q) << q
        denominator = 1
    else:
        numerator = _integer_sqrt_of_frac_rto(n << -2 * q, m)
        denominator = 1 << -q
    return numerator / denominator


def mean(data):
    T, total, n = _sum(data)
    if n < 1:
        raise StatisticsError("mean requires at least one data point")
    return _convert(total / n, T)


def fmean(data, weights=None):
    if weights is None:
        data = list(data)
        n = len(data)
        total = math.fsum(data)
        if not n:
            raise StatisticsError("fmean requires at least one data point")
        return total / n
    data = list(data)
    weights = list(weights)
    num = math.sumprod(data, weights)
    if len(data) != len(weights):
        raise ValueError("data and weights must be the same length")
    den = math.fsum(weights)
    if not den:
        raise ZeroDivisionError("sum of weights must be non-zero")
    return num / den


def geometric_mean(data):
    n = 0
    found_zero = False
    logs = []
    for x in data:
        n += 1
        if x > 0.0 or math.isnan(x):
            logs.append(math.log(x))
        elif x == 0.0:
            found_zero = True
        else:
            raise StatisticsError("No negative inputs allowed", x)
    total = math.fsum(logs)
    if not n:
        raise StatisticsError("Must have a non-empty dataset")
    if math.isnan(total):
        return math.nan
    if found_zero:
        if total == math.inf:
            return math.nan
        return 0.0
    return math.exp(total / n)


def harmonic_mean(data, weights=None):
    data = list(data)
    errmsg = "harmonic mean does not support negative values"
    n = len(data)
    if n < 1:
        raise StatisticsError("harmonic_mean requires at least one data point")
    elif n == 1 and weights is None:
        x = data[0]
        if isinstance(x, (int, float, Fraction)):
            if x < 0:
                raise StatisticsError(errmsg)
            return x
        raise TypeError("unsupported type")
    if weights is None:
        weights = [1] * n
        sum_weights = n
    else:
        weights = list(weights)
        if len(weights) != n:
            raise StatisticsError("Number of weights does not match data size")
        _, sum_weights, _ = _sum(_fail_neg(weights, errmsg))
    try:
        data = _fail_neg(data, errmsg)
        terms = []
        for w, x in zip(weights, data):
            if w:
                terms.append(w / x)
            else:
                terms.append(0)
        T, total, count = _sum(terms)
    except ZeroDivisionError:
        return 0
    if total <= 0:
        raise StatisticsError("Weighted sum must be positive")
    return _convert(sum_weights / total, T)


def median(data):
    data = sorted(data)
    n = len(data)
    if n == 0:
        raise StatisticsError("no median for empty data")
    if n % 2 == 1:
        return data[n // 2]
    else:
        i = n // 2
        return (data[i - 1] + data[i]) / 2


def median_low(data):
    data = sorted(data)
    n = len(data)
    if n == 0:
        raise StatisticsError("no median for empty data")
    if n % 2 == 1:
        return data[n // 2]
    else:
        return data[n // 2 - 1]


def median_high(data):
    data = sorted(data)
    n = len(data)
    if n == 0:
        raise StatisticsError("no median for empty data")
    return data[n // 2]


def median_grouped(data, interval=1.0):
    data = sorted(data)
    n = len(data)
    if not n:
        raise StatisticsError("no median for empty data")
    elif n == 1:
        return data[0]
    x = data[n // 2]
    try:
        L = x - interval / 2.0
    except TypeError:
        L = float(x) - float(interval) / 2.0
    l1 = bisect_left(data, x)
    l2 = bisect_right(data, x)
    cf = l1
    f = l2 - l1
    return L + interval * (n / 2 - cf) / f


def mode(data):
    pairs = Counter(list(data)).most_common(1)
    if not pairs:
        raise StatisticsError("no mode for empty data")
    return pairs[0][0]


def multimode(data):
    counts = Counter(list(data))
    if not counts:
        return []
    maxcount = max(counts.values())
    return [value for value, count in counts.items() if count == maxcount]


def quantiles(data, *, n=4, method="exclusive"):
    if n < 1:
        raise StatisticsError("n must be at least 1")
    data = sorted(data)
    ld = len(data)
    if ld < 2:
        if ld == 1:
            return data * (n - 1)
        raise StatisticsError("must have at least one data point")
    if method == "inclusive":
        m = ld - 1
        result = []
        for i in range(1, n):
            j, delta = divmod(i * m, n)
            interpolated = (data[j] * (n - delta) + data[j + 1] * delta) / n
            result.append(interpolated)
        return result
    if method == "exclusive":
        m = ld + 1
        result = []
        for i in range(1, n):
            j = i * m // n
            if j < 1:
                j = 1
            elif j > ld - 1:
                j = ld - 1
            delta = i * m - j * n
            interpolated = (data[j - 1] * (n - delta) + data[j] * delta) / n
            result.append(interpolated)
        return result
    raise ValueError("Unknown method: " + repr(method))


def _ss(data, c=None):
    if c is not None:
        sq = []
        for x in data:
            d = x - c
            sq.append(d * d)
        T, ssd, count = _sum(sq)
        return (T, ssd, c, count)
    count = 0
    types = set()
    sx_partials = {}
    sxx_partials = {}
    for v in data:
        types.add(type(v))
        pair = _exact_ratio(v)
        n, d = pair[0], pair[1]
        count += 1
        sx_partials[d] = sx_partials.get(d, 0) + n
        sxx_partials[d] = sxx_partials.get(d, 0) + n * n
    if not count:
        ssd = Fraction(0)
        c = Fraction(0)
    elif None in sx_partials:
        ssd = sx_partials[None]
        c = ssd
    else:
        sx = Fraction(0)
        for d, n in sx_partials.items():
            sx = sx + Fraction(n, d)
        sxx = Fraction(0)
        for d, n in sxx_partials.items():
            sxx = sxx + Fraction(n, d * d)
        ssd = (count * sxx - sx * sx) / count
        c = sx / count
    T = int
    for t in types:
        T = _coerce(T, t)
    return (T, ssd, c, count)


def variance(data, xbar=None):
    T, ss, c, n = _ss(list(data), xbar)
    if n < 2:
        raise StatisticsError("variance requires at least two data points")
    return _convert(ss / (n - 1), T)


def pvariance(data, mu=None):
    T, ss, c, n = _ss(list(data), mu)
    if n < 1:
        raise StatisticsError("pvariance requires at least one data point")
    return _convert(ss / n, T)


def stdev(data, xbar=None):
    T, ss, c, n = _ss(list(data), xbar)
    if n < 2:
        raise StatisticsError("stdev requires at least two data points")
    mss = ss / (n - 1)
    return _float_sqrt_of_frac(mss.numerator, mss.denominator)


def pstdev(data, mu=None):
    T, ss, c, n = _ss(list(data), mu)
    if n < 1:
        raise StatisticsError("pstdev requires at least one data point")
    mss = ss / n
    return _float_sqrt_of_frac(mss.numerator, mss.denominator)


def covariance(x, y, /):
    n = len(x)
    if len(y) != n:
        raise StatisticsError("covariance requires that both inputs have same number of data points")
    if n < 2:
        raise StatisticsError("covariance requires at least two data points")
    xbar = math.fsum(x) / n
    ybar = math.fsum(y) / n
    sxy = math.sumprod([xi - xbar for xi in x], [yi - ybar for yi in y])
    return sxy / (n - 1)


def _rank(data):
    order = sorted(range(len(data)), key=lambda i: data[i])
    ranks = [0.0] * len(data)
    i = 0
    n = len(data)
    while i < n:
        j = i
        while j + 1 < n and data[order[j + 1]] == data[order[i]]:
            j += 1
        r = (i + j) / 2 + 1
        for k in range(i, j + 1):
            ranks[order[k]] = r
        i = j + 1
    return ranks


def correlation(x, y, /, *, method="linear"):
    n = len(x)
    if len(y) != n:
        raise StatisticsError("correlation requires that both inputs have same number of data points")
    if n < 2:
        raise StatisticsError("correlation requires at least two data points")
    if method != "linear" and method != "ranked":
        raise ValueError("Unknown method: " + repr(method))
    if method == "ranked":
        x = _rank(x)
        y = _rank(y)
        xbar = (n + 1) / 2
        ybar = xbar
    else:
        xbar = math.fsum(x) / n
        ybar = math.fsum(y) / n
    xdev = [xi - xbar for xi in x]
    ydev = [yi - ybar for yi in y]
    sxy = math.sumprod(xdev, ydev)
    sxx = math.sumprod(xdev, xdev)
    syy = math.sumprod(ydev, ydev)
    try:
        return sxy / math.sqrt(sxx * syy)
    except ZeroDivisionError:
        raise StatisticsError("at least one of the inputs is constant") from None


LinearRegression = namedtuple("LinearRegression", ("slope", "intercept"))


def linear_regression(x, y, /, *, proportional=False):
    n = len(x)
    if len(y) != n:
        raise StatisticsError("linear regression requires that both inputs have same number of data points")
    if n < 2:
        raise StatisticsError("linear regression requires at least two data points")
    if proportional:
        sxy = math.sumprod(x, y)
        sxx = math.sumprod(x, x)
        slope = sxy / sxx
        intercept = 0.0
    else:
        xbar = math.fsum(x) / n
        ybar = math.fsum(y) / n
        xdev = [xi - xbar for xi in x]
        ydev = [yi - ybar for yi in y]
        sxy = math.sumprod(xdev, ydev)
        sxx = math.sumprod(xdev, xdev)
        try:
            slope = sxy / sxx
        except ZeroDivisionError:
            raise StatisticsError("x is constant") from None
        intercept = ybar - slope * xbar
    return LinearRegression(slope=slope, intercept=intercept)


_SQRT2 = math.sqrt(2.0)


class NormalDist:
    def __init__(self, mu=0.0, sigma=1.0):
        if sigma < 0.0:
            raise StatisticsError("sigma must be non-negative")
        self._mu = float(mu)
        self._sigma = float(sigma)

    @property
    def mean(self):
        return self._mu

    @property
    def median(self):
        return self._mu

    @property
    def mode(self):
        return self._mu

    @property
    def stdev(self):
        return self._sigma

    @property
    def variance(self):
        return self._sigma * self._sigma

    @classmethod
    def from_samples(cls, data):
        return cls(fmean(data), stdev(data))

    def pdf(self, x):
        variance = self._sigma * self._sigma
        if not variance:
            raise StatisticsError("pdf() not defined when sigma is zero")
        diff = x - self._mu
        return math.exp(diff * diff / (-2.0 * variance)) / math.sqrt(math.tau * variance)

    def cdf(self, x):
        if not self._sigma:
            raise StatisticsError("cdf() not defined when sigma is zero")
        return 0.5 * (1.0 + math.erf((x - self._mu) / (self._sigma * _SQRT2)))

    def zscore(self, x):
        if not self._sigma:
            raise StatisticsError("zscore() not defined when sigma is zero")
        return (x - self._mu) / self._sigma

    def __add__(x1, x2):
        if isinstance(x2, NormalDist):
            return NormalDist(x1._mu + x2._mu, math.hypot(x1._sigma, x2._sigma))
        return NormalDist(x1._mu + x2, x1._sigma)

    def __sub__(x1, x2):
        if isinstance(x2, NormalDist):
            return NormalDist(x1._mu - x2._mu, math.hypot(x1._sigma, x2._sigma))
        return NormalDist(x1._mu - x2, x1._sigma)

    def __mul__(x1, x2):
        return NormalDist(x1._mu * x2, x1._sigma * math.fabs(x2))

    def __truediv__(x1, x2):
        return NormalDist(x1._mu / x2, x1._sigma / math.fabs(x2))

    def __pos__(x1):
        return NormalDist(x1._mu, x1._sigma)

    def __neg__(x1):
        return NormalDist(-x1._mu, x1._sigma)

    def __radd__(x1, x2):
        return x1 + x2

    def __rsub__(x1, x2):
        return -(x1 - x2)

    def __rmul__(x1, x2):
        return x1 * x2

    def __eq__(x1, x2):
        if not isinstance(x2, NormalDist):
            return NotImplemented
        return x1._mu == x2._mu and x1._sigma == x2._sigma

    def __hash__(self):
        return hash((self._mu, self._sigma))

    def __repr__(self):
        return "NormalDist(mu=" + repr(self._mu) + ", sigma=" + repr(self._sigma) + ")"
