"""fractions: rational numbers."""

import math
import re

_RATIONAL_FORMAT = re.compile(r"\A\s*(?P<sign>[-+]?)(?=\d|\.\d)(?P<num>\d*|\d+(_\d+)*)(?:(?:/(?P<denom>\d+(_\d+)*))?|(?:\.(?P<decimal>\d*|\d+(_\d+)*))?(?:E(?P<exp>[-+]?\d+(_\d+)*))?)\s*\Z", re.IGNORECASE)


def _isint(x):
    return isinstance(x, int)


def _is_frac(x):
    return isinstance(x, Fraction)


class Fraction:
    def __init__(self, numerator=0, denominator=None):
        if denominator is None:
            if _isint(numerator):
                self._numerator = int(numerator)
                self._denominator = 1
                return
            if _is_frac(numerator):
                self._numerator = numerator._numerator
                self._denominator = numerator._denominator
                return
            if isinstance(numerator, float):
                n, d = numerator.as_integer_ratio()
                self._numerator = n
                self._denominator = d
                return
            if hasattr(numerator, "as_integer_ratio"):
                pair = numerator.as_integer_ratio()
                self._numerator = pair[0]
                self._denominator = pair[1]
                return
            if isinstance(numerator, str):
                m = _RATIONAL_FORMAT.match(numerator)
                if m is None:
                    raise ValueError("Invalid literal for Fraction: " + repr(numerator))
                num = int(m.group("num") or "0")
                denom = m.group("denom")
                if denom:
                    denom = int(denom)
                else:
                    denom = 1
                    decimal = m.group("decimal")
                    if decimal:
                        decimal = decimal.replace("_", "")
                        scale = 10 ** len(decimal)
                        num = num * scale + int(decimal)
                        denom *= scale
                    exp = m.group("exp")
                    if exp:
                        exp = int(exp)
                        if exp >= 0:
                            num *= 10 ** exp
                        else:
                            denom *= 10 ** -exp
                if denom == 0:
                    raise ZeroDivisionError("Fraction(" + str(num) + ", 0)")
                if m.group("sign") == "-":
                    num = -num
                g = math.gcd(num, denom)
                self._numerator = num // g
                self._denominator = denom // g
                return
            raise TypeError("argument should be a string or a Rational instance")
        if _isint(numerator) and _isint(denominator):
            pass
        elif _is_frac(numerator) or _is_frac(denominator) or hasattr(numerator, "numerator") and hasattr(denominator, "numerator"):
            na, da = numerator.numerator, numerator.denominator
            nb, db = denominator.numerator, denominator.denominator
            numerator, denominator = na * db, da * nb
        else:
            raise TypeError("both arguments should be Rational instances")
        if denominator == 0:
            raise ZeroDivisionError("Fraction(" + str(numerator) + ", 0)")
        g = math.gcd(numerator, denominator)
        if denominator < 0:
            g = -g
        self._numerator = numerator // g
        self._denominator = denominator // g

    @classmethod
    def from_float(cls, f):
        if _isint(f):
            return cls(f)
        if not isinstance(f, float):
            raise TypeError("Fraction.from_float() only takes floats, not " + repr(f) + " (" + type(f).__name__ + ")")
        return cls(*f.as_integer_ratio())

    @classmethod
    def from_decimal(cls, dec):
        if _isint(dec):
            return cls(dec)
        return cls(*dec.as_integer_ratio())

    def as_integer_ratio(self):
        return (self._numerator, self._denominator)

    def is_integer(self):
        return self._denominator == 1

    def limit_denominator(self, max_denominator=1000000):
        if max_denominator < 1:
            raise ValueError("max_denominator should be at least 1")
        if self._denominator <= max_denominator:
            return Fraction(self)
        p0, q0, p1, q1 = 0, 1, 1, 0
        n, d = self._numerator, self._denominator
        while True:
            a = n // d
            q2 = q0 + a * q1
            if q2 > max_denominator:
                break
            p0, q0, p1, q1 = p1, q1, p0 + a * p1, q2
            n, d = d, n - a * d
        k = (max_denominator - q0) // q1
        if 2 * d * (q0 + k * q1) <= self._denominator:
            return Fraction(p1, q1)
        return Fraction(p0 + k * p1, q0 + k * q1)

    @property
    def numerator(self):
        return self._numerator

    @property
    def denominator(self):
        return self._denominator

    def __repr__(self):
        return "Fraction(" + str(self._numerator) + ", " + str(self._denominator) + ")"

    def __str__(self):
        if self._denominator == 1:
            return str(self._numerator)
        return str(self._numerator) + "/" + str(self._denominator)

    def __format__(self, spec):
        if not spec:
            return str(self)
        m = re.match(r"\A(?:(?P<fill>.)?(?P<align>[<>=^]))?(?P<sign>[-+ ]?)(?P<no_neg_zero>z?)(?P<alt>#?)(?P<zeropad>0(?=\d))?(?P<minimumwidth>\d+)?(?P<thousands_sep>[,_])?(?:\.(?P<precision>\d+))?(?P<presentation_type>[eEfFgG%])\Z", spec)
        if m is None:
            if spec[-1:] in "eEfFgG%":
                raise ValueError("Invalid format specifier '" + spec + "' for object of type 'Fraction'")
            return format(str(self), spec)
        presentation = m.group("presentation_type")
        precision = m.group("precision")
        if precision is None:
            precision = 6
        else:
            precision = int(precision)
        if presentation == "%":
            value = self * 100
        else:
            value = self
        sign = "-" if value._numerator < 0 else ""
        if presentation in "fF%":
            scaled = abs(value._numerator) * 10 ** precision
            q, r = divmod(scaled, value._denominator)
            if 2 * r >= value._denominator and (2 * r > value._denominator or q % 2 == 1):
                q += 1
            digits = str(q).rjust(precision + 1, "0")
            body = digits[:len(digits) - precision] + ("." + digits[len(digits) - precision:] if precision else "")
            if presentation == "%":
                body += "%"
        else:
            return format(float(self), spec)
        if sign == "" and m.group("sign") in ("+", " "):
            sign = m.group("sign")
        text = sign + body
        width = m.group("minimumwidth")
        if width:
            width = int(width)
            fill = m.group("fill") or " "
            align = m.group("align") or ">"
            if m.group("zeropad"):
                text = sign + body.rjust(width - len(sign), "0")
            elif align == "<":
                text = text.ljust(width, fill)
            elif align == "^":
                text = text.center(width, fill)
            else:
                text = text.rjust(width, fill)
        return text

    def __add__(self, other):
        if _is_frac(other) or _isint(other):
            o = Fraction(other)
            na, da = self._numerator, self._denominator
            nb, db = o._numerator, o._denominator
            g = math.gcd(da, db)
            if g == 1:
                return Fraction._from_coprime(na * db + da * nb, da * db)
            s = da // g
            t = na * (db // g) + nb * s
            g2 = math.gcd(t, g)
            if g2 == 1:
                return Fraction._from_coprime(t, s * db)
            return Fraction._from_coprime(t // g2, s * (db // g2))
        if isinstance(other, float):
            return float(self) + other
        if isinstance(other, complex):
            return complex(self) + other
        return NotImplemented

    def __radd__(self, other):
        return self + other

    def __sub__(self, other):
        if _is_frac(other) or _isint(other):
            o = Fraction(other)
            na, da = self._numerator, self._denominator
            nb, db = o._numerator, o._denominator
            g = math.gcd(da, db)
            if g == 1:
                return Fraction._from_coprime(na * db - da * nb, da * db)
            s = da // g
            t = na * (db // g) - nb * s
            g2 = math.gcd(t, g)
            if g2 == 1:
                return Fraction._from_coprime(t, s * db)
            return Fraction._from_coprime(t // g2, s * (db // g2))
        if isinstance(other, float):
            return float(self) - other
        if isinstance(other, complex):
            return complex(self) - other
        return NotImplemented

    def __rsub__(self, other):
        if _is_frac(other) or _isint(other):
            return Fraction(other) - self
        if isinstance(other, float):
            return other - float(self)
        if isinstance(other, complex):
            return other - complex(self)
        return NotImplemented

    def __mul__(self, other):
        if _is_frac(other) or _isint(other):
            o = Fraction(other)
            na, da = self._numerator, self._denominator
            nb, db = o._numerator, o._denominator
            g1 = math.gcd(na, db)
            if g1 > 1:
                na //= g1
                db //= g1
            g2 = math.gcd(nb, da)
            if g2 > 1:
                nb //= g2
                da //= g2
            return Fraction._from_coprime(na * nb, db * da)
        if isinstance(other, float):
            return float(self) * other
        if isinstance(other, complex):
            return complex(self) * other
        return NotImplemented

    def __rmul__(self, other):
        return self * other

    def __truediv__(self, other):
        if _is_frac(other) or _isint(other):
            o = Fraction(other)
            na, da = self._numerator, self._denominator
            nb, db = o._numerator, o._denominator
            if nb == 0:
                raise ZeroDivisionError("Fraction(" + str(da) + ", 0)")
            g1 = math.gcd(na, nb)
            if g1 > 1:
                na //= g1
                nb //= g1
            g2 = math.gcd(db, da)
            if g2 > 1:
                da //= g2
                db //= g2
            n, d = na * db, nb * da
            if d < 0:
                n, d = -n, -d
            return Fraction._from_coprime(n, d)
        if isinstance(other, float):
            return float(self) / other
        if isinstance(other, complex):
            return complex(self) / other
        return NotImplemented

    def __rtruediv__(self, other):
        if _is_frac(other) or _isint(other):
            return Fraction(other) / self
        if isinstance(other, float):
            return other / float(self)
        if isinstance(other, complex):
            return other / complex(self)
        return NotImplemented

    def __floordiv__(self, other):
        if _is_frac(other) or _isint(other):
            o = Fraction(other)
            return (self._numerator * o._denominator) // (self._denominator * o._numerator)
        if isinstance(other, float):
            return float(self) // other
        return NotImplemented

    def __rfloordiv__(self, other):
        if _is_frac(other) or _isint(other):
            o = Fraction(other)
            return (o._numerator * self._denominator) // (o._denominator * self._numerator)
        if isinstance(other, float):
            return other // float(self)
        return NotImplemented

    def __mod__(self, other):
        if _is_frac(other) or _isint(other):
            o = Fraction(other)
            da, db = self._denominator, o._denominator
            return Fraction((self._numerator * db) % (o._numerator * da), da * db)
        if isinstance(other, float):
            return float(self) % other
        return NotImplemented

    def __rmod__(self, other):
        if _is_frac(other) or _isint(other):
            return Fraction(other) % self
        if isinstance(other, float):
            return other % float(self)
        return NotImplemented

    def __divmod__(self, other):
        if _is_frac(other) or _isint(other):
            o = Fraction(other)
            da, db = self._denominator, o._denominator
            div, n_mod = divmod(self._numerator * db, da * o._numerator)
            return div, Fraction(n_mod, da * db)
        if isinstance(other, float):
            return divmod(float(self), other)
        return NotImplemented

    def __rdivmod__(self, other):
        if _is_frac(other) or _isint(other):
            return divmod(Fraction(other), self)
        if isinstance(other, float):
            return divmod(other, float(self))
        return NotImplemented

    def __pow__(self, b):
        if _is_frac(b) or _isint(b):
            exp = Fraction(b)
            if exp._denominator == 1:
                power = exp._numerator
                if power >= 0:
                    return Fraction._from_coprime(self._numerator ** power, self._denominator ** power)
                elif self._numerator > 0:
                    return Fraction._from_coprime(self._denominator ** -power, self._numerator ** -power)
                elif self._numerator == 0:
                    raise ZeroDivisionError("Fraction(" + str(self._denominator ** -power) + ", 0)")
                else:
                    return Fraction._from_coprime((-self._denominator) ** -power, (-self._numerator) ** -power)
            return float(self) ** float(exp)
        if isinstance(b, float):
            return float(self) ** b
        if isinstance(b, complex):
            return complex(self) ** b
        return NotImplemented

    def __rpow__(self, a):
        if self._denominator == 1 and self._numerator >= 0:
            return a ** self._numerator
        if _is_frac(a) or _isint(a):
            return Fraction(a) ** self
        if isinstance(a, float):
            return a ** float(self)
        if isinstance(a, complex):
            return a ** complex(self)
        return NotImplemented

    @staticmethod
    def _from_coprime(n, d):
        obj = Fraction(0)
        obj._numerator = n
        obj._denominator = d
        return obj

    def __pos__(self):
        return Fraction._from_coprime(self._numerator, self._denominator)

    def __neg__(self):
        return Fraction._from_coprime(-self._numerator, self._denominator)

    def __abs__(self):
        return Fraction._from_coprime(abs(self._numerator), self._denominator)

    def conjugate(self):
        return self

    def __int__(self):
        if self._numerator < 0:
            return -(-self._numerator // self._denominator)
        return self._numerator // self._denominator

    def __trunc__(self):
        return int(self)

    def __floor__(self):
        return self._numerator // self._denominator

    def __ceil__(self):
        return -(-self._numerator // self._denominator)

    def __float__(self):
        return self._numerator / self._denominator

    def __complex__(self):
        return complex(float(self))

    def __round__(self, ndigits=None):
        if ndigits is None:
            floor, remainder = divmod(self._numerator, self._denominator)
            if remainder * 2 < self._denominator:
                return floor
            elif remainder * 2 > self._denominator:
                return floor + 1
            elif floor % 2 == 0:
                return floor
            else:
                return floor + 1
        shift = 10 ** abs(ndigits)
        if ndigits > 0:
            return Fraction(round(self * shift), shift)
        return Fraction(round(self / shift) * shift)

    def __hash__(self):
        modulus = 2305843009213693951
        dinv = pow(self._denominator, -1, modulus) if self._denominator % modulus != 0 else 0
        if self._denominator % modulus == 0:
            hash_ = 314159265358979323
        else:
            hash_ = abs(self._numerator) * dinv % modulus
        result = hash_
        if self._numerator < 0:
            result = -hash_
        if result == -1:
            return -2
        return result

    def __eq__(self, b):
        if _isint(b):
            return self._numerator == b and self._denominator == 1
        if _is_frac(b):
            return self._numerator == b._numerator and self._denominator == b._denominator
        if isinstance(b, complex):
            if b.imag == 0:
                b = b.real
            else:
                return False
        if isinstance(b, float):
            if math.isnan(b) or math.isinf(b):
                return 0.0 == b
            return self == Fraction.from_float(b)
        return NotImplemented

    def __lt__(self, other):
        if _is_frac(other) or _isint(other) or isinstance(other, float):
            if isinstance(other, float) and (math.isnan(other) or math.isinf(other)):
                return 0.0 < other
            o = Fraction(other)
            return self._numerator * o._denominator < o._numerator * self._denominator
        return NotImplemented

    def __gt__(self, other):
        if _is_frac(other) or _isint(other) or isinstance(other, float):
            if isinstance(other, float) and (math.isnan(other) or math.isinf(other)):
                return 0.0 > other
            o = Fraction(other)
            return self._numerator * o._denominator > o._numerator * self._denominator
        return NotImplemented

    def __le__(self, other):
        if _is_frac(other) or _isint(other) or isinstance(other, float):
            if isinstance(other, float) and (math.isnan(other) or math.isinf(other)):
                return 0.0 <= other
            o = Fraction(other)
            return self._numerator * o._denominator <= o._numerator * self._denominator
        return NotImplemented

    def __ge__(self, other):
        if _is_frac(other) or _isint(other) or isinstance(other, float):
            if isinstance(other, float) and (math.isnan(other) or math.isinf(other)):
                return 0.0 >= other
            o = Fraction(other)
            return self._numerator * o._denominator >= o._numerator * self._denominator
        return NotImplemented

    def __bool__(self):
        return self._numerator != 0


import numbers as _numbers
_numbers.Rational.register(Fraction)
