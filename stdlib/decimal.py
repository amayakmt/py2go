"""decimal: decimal floating point arithmetic, after the General Decimal Arithmetic specification."""

import math as _math
import re as _re
import sys as _sys
from collections import namedtuple as _namedtuple
from fractions import Fraction as _Fraction

ROUND_DOWN = "ROUND_DOWN"
ROUND_HALF_UP = "ROUND_HALF_UP"
ROUND_HALF_EVEN = "ROUND_HALF_EVEN"
ROUND_CEILING = "ROUND_CEILING"
ROUND_FLOOR = "ROUND_FLOOR"
ROUND_UP = "ROUND_UP"
ROUND_HALF_DOWN = "ROUND_HALF_DOWN"
ROUND_05UP = "ROUND_05UP"

MAX_PREC = 999999999999999999
MAX_EMAX = 999999999999999999
MIN_EMIN = -999999999999999999
MIN_ETINY = MIN_EMIN - (MAX_PREC - 1)
HAVE_THREADS = True
HAVE_CONTEXTVAR = True
__version__ = "1.70"
__libmpdec_version__ = "2.5.1"


class DecimalException(ArithmeticError):
    def handle(self, context, *args):
        pass


class Clamped(DecimalException):
    pass


class InvalidOperation(DecimalException):
    pass


class ConversionSyntax(InvalidOperation):
    pass


class DivisionByZero(DecimalException, ZeroDivisionError):
    pass


class DivisionImpossible(InvalidOperation):
    pass


class DivisionUndefined(InvalidOperation, ZeroDivisionError):
    pass


class Inexact(DecimalException):
    pass


class InvalidContext(InvalidOperation):
    pass


class Rounded(DecimalException):
    pass


class Subnormal(DecimalException):
    pass


class Overflow(Inexact, Rounded):
    pass


class Underflow(Inexact, Rounded, Subnormal):
    pass


class FloatOperation(DecimalException, TypeError):
    pass


_signals = [InvalidOperation, FloatOperation, DivisionByZero, Overflow, Underflow, Subnormal, Inexact, Rounded, Clamped]
_repr_order = [Clamped, InvalidOperation, DivisionByZero, Inexact, FloatOperation, Overflow, Rounded, Subnormal, Underflow]


def _signal_name(s):
    return s.__name__


class _ContextState:
    def __init__(self):
        self.current = {}
        self.default = None


_cstate = _ContextState()


def _ident():
    return _native("thread.get_ident")


def getcontext():
    ident = _ident()
    c = _cstate.current.get(ident)
    if c is None:
        c = Context()
        _cstate.current[ident] = c
    return c


def setcontext(context):
    if context is DefaultContext or context is BasicContext or context is ExtendedContext:
        context = context.copy()
        context.clear_flags()
    _cstate.current[_ident()] = context


def _getc(context):
    if context is None:
        return getcontext()
    return context


class _ContextManager:
    def __init__(self, new_context):
        self.new_context = new_context.copy()
        self.saved_context = None

    def __enter__(self):
        self.saved_context = getcontext()
        setcontext(self.new_context)
        return self.new_context

    def __exit__(self, t, v, tb):
        setcontext(self.saved_context)


def localcontext(ctx=None, **kwargs):
    if ctx is None:
        ctx = getcontext()
    ctx_manager = _ContextManager(ctx)
    for key, value in kwargs.items():
        setattr(ctx_manager.new_context, key, value)
    return ctx_manager


DecimalTuple = _namedtuple("DecimalTuple", "sign digits exponent")

_parser = _re.compile(r"""\A
        (?P<sign>[-+])?
        (
            (?=\d|\.\d)
            (?P<int>\d*)
            (\.(?P<frac>\d*))?
            (E(?P<exp>[-+]?\d+))?
        |
            Inf(inity)?
        |
            (?P<signal>s)?
            NaN
            (?P<diag>\d*)
        )
        \Z
""", _re.VERBOSE | _re.IGNORECASE)


def _dec_from_triple(sign, coefficient, exponent, special=False):
    self = Decimal(0)
    self._sign = sign
    self._int = coefficient
    self._exp = exponent
    self._is_special = special
    return self


def _nbits(n):
    return n.bit_length()


class Decimal:
    def __init__(self, value="0", context=None):
        if context is not None and not isinstance(context, Context):
            raise TypeError("optional argument must be a context")
        self._sign = 0
        self._int = "0"
        self._exp = 0
        self._is_special = False
        if isinstance(value, str):
            m = _parser.match(value.strip().replace("_", ""))
            if m is None:
                if context is None:
                    context = getcontext()
                self._set_nan_syntax(context, value)
                return
            if m.group("sign") == "-":
                self._sign = 1
            else:
                self._sign = 0
            intpart = m.group("int")
            if intpart is not None:
                fracpart = m.group("frac") or ""
                exp = int(m.group("exp") or "0")
                self._int = str(int(intpart + fracpart))
                self._exp = exp - len(fracpart)
                self._is_special = False
            else:
                diag = m.group("diag")
                if diag is not None:
                    self._int = str(int(diag or "0")).lstrip("0")
                    if m.group("signal"):
                        self._exp = "N"
                    else:
                        self._exp = "n"
                else:
                    self._int = "0"
                    self._exp = "F"
                self._is_special = True
            return
        if isinstance(value, int):
            ival = int(value)
            if ival >= 0:
                self._sign = 0
            else:
                self._sign = 1
            self._exp = 0
            self._int = str(abs(ival))
            return
        if isinstance(value, Decimal):
            self._exp = value._exp
            self._sign = value._sign
            self._int = value._int
            self._is_special = value._is_special
            return
        if isinstance(value, (list, tuple)):
            if len(value) != 3:
                raise ValueError("argument must be a sequence of length 3")
            if not (isinstance(value[0], int) and value[0] in (0, 1)):
                raise ValueError("sign must be an integer with the value 0 or 1")
            self._sign = value[0]
            if value[2] == "F":
                self._int = "0"
                self._exp = value[2]
                self._is_special = True
            else:
                digits = []
                for digit in value[1]:
                    if isinstance(digit, int) and 0 <= digit <= 9:
                        if digits or digit != 0:
                            digits.append(digit)
                    else:
                        raise ValueError("coefficient must be a tuple of digits")
                if value[2] in ("n", "N"):
                    self._int = "".join([str(d) for d in digits])
                    self._exp = value[2]
                    self._is_special = True
                elif isinstance(value[2], int):
                    self._int = "".join([str(d) for d in (digits or [0])])
                    self._exp = value[2]
                    self._is_special = False
                else:
                    raise ValueError("exponent must be an integer")
            return
        if isinstance(value, float):
            if context is None:
                context = getcontext()
            if context.traps[FloatOperation]:
                context._raise_error(FloatOperation, "strict semantics for mixing floats and Decimals are enabled")
            else:
                context.flags[FloatOperation] = True
            d = Decimal._from_float_exact(value)
            self._exp = d._exp
            self._sign = d._sign
            self._int = d._int
            self._is_special = d._is_special
            return
        raise TypeError("conversion from " + type(value).__name__ + " to Decimal is not supported")

    def _set_nan_syntax(self, context, value):
        self._sign = 0
        self._int = "0"
        self._exp = "n"
        self._is_special = True
        context._raise_error(ConversionSyntax, "invalid literal for Decimal: " + repr(value))

    @classmethod
    def _from_float_exact(cls, f):
        if isinstance(f, int):
            return cls(f)
        if _math.isinf(f) or _math.isnan(f):
            return cls(repr(f))
        if _math.copysign(1.0, f) == 1.0:
            sign = 0
        else:
            sign = 1
        n, d = abs(f).as_integer_ratio()
        k = d.bit_length() - 1
        result = _dec_from_triple(sign, str(n * 5 ** k), -k)
        return result

    @classmethod
    def from_float(cls, f):
        if isinstance(f, int):
            return cls(f)
        if not isinstance(f, float):
            raise TypeError("argument must be int or float.")
        result = cls._from_float_exact(f)
        if cls is not Decimal:
            result = cls(result)
        return result

    def _isnan(self):
        if self._is_special:
            exp = self._exp
            if exp == "n":
                return 1
            elif exp == "N":
                return 2
        return 0

    def _isinfinity(self):
        if self._exp == "F":
            if self._sign:
                return -1
            return 1
        return 0

    def _check_nans(self, other=None, context=None):
        self_is_nan = self._isnan()
        if other is None:
            other_is_nan = False
        else:
            other_is_nan = other._isnan()
        if self_is_nan or other_is_nan:
            if context is None:
                context = getcontext()
            if self_is_nan == 2:
                return context._raise_error(InvalidOperation, "sNaN", self)
            if other_is_nan == 2:
                return context._raise_error(InvalidOperation, "sNaN", other)
            if self_is_nan:
                return self._fix_nan(context)
            return other._fix_nan(context)
        return 0

    def _compare_check_nans(self, other, context):
        if context is None:
            context = getcontext()
        if self._is_special or other._is_special:
            if self.is_snan():
                return context._raise_error(InvalidOperation, "comparison involving sNaN", self)
            elif other.is_snan():
                return context._raise_error(InvalidOperation, "comparison involving sNaN", other)
            elif self.is_qnan():
                return context._raise_error(InvalidOperation, "comparison involving NaN", self)
            elif other.is_qnan():
                return context._raise_error(InvalidOperation, "comparison involving NaN", other)
        return 0

    def __bool__(self):
        return self._is_special or self._int != "0"

    def _cmp(self, other):
        if self._is_special or other._is_special:
            self_inf = self._isinfinity()
            other_inf = other._isinfinity()
            if self_inf == other_inf:
                return 0
            elif self_inf < other_inf:
                return -1
            else:
                return 1
        if not self:
            if not other:
                return 0
            else:
                return -((-1) ** other._sign)
        if not other:
            return (-1) ** self._sign
        if other._sign < self._sign:
            return -1
        if self._sign < other._sign:
            return 1
        self_adjusted = self.adjusted()
        other_adjusted = other.adjusted()
        if self_adjusted == other_adjusted:
            self_padded = self._int + "0" * (self._exp - other._exp)
            other_padded = other._int + "0" * (other._exp - self._exp)
            if self_padded == other_padded:
                return 0
            elif self_padded < other_padded:
                return -((-1) ** self._sign)
            else:
                return (-1) ** self._sign
        elif self_adjusted > other_adjusted:
            return (-1) ** self._sign
        else:
            return -((-1) ** self._sign)

    def __eq__(self, other):
        pair = _cmp_operands(self, other, True)
        if pair is None:
            return NotImplemented
        a, o = pair
        if a._check_nans(o, None):
            return False
        return a._cmp(o) == 0

    def __lt__(self, other):
        pair = _cmp_operands(self, other, False)
        if pair is None:
            return NotImplemented
        a, o = pair
        ans = a._compare_check_nans(o, None)
        if ans:
            return False
        return a._cmp(o) < 0

    def __le__(self, other):
        pair = _cmp_operands(self, other, False)
        if pair is None:
            return NotImplemented
        a, o = pair
        ans = a._compare_check_nans(o, None)
        if ans:
            return False
        return a._cmp(o) <= 0

    def __gt__(self, other):
        pair = _cmp_operands(self, other, False)
        if pair is None:
            return NotImplemented
        a, o = pair
        ans = a._compare_check_nans(o, None)
        if ans:
            return False
        return a._cmp(o) > 0

    def __ge__(self, other):
        pair = _cmp_operands(self, other, False)
        if pair is None:
            return NotImplemented
        a, o = pair
        ans = a._compare_check_nans(o, None)
        if ans:
            return False
        return a._cmp(o) >= 0

    def compare(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        if self._is_special or other._is_special:
            ans = self._check_nans(other, context)
            if ans:
                return ans
        return Decimal(self._cmp(other))

    def __hash__(self):
        if self._is_special:
            if self.is_snan():
                raise TypeError("Cannot hash a signaling NaN value")
            elif self.is_nan():
                return 0
            else:
                if self._sign:
                    return -314159
                return 314159
        modulus = 2305843009213693951
        if self._exp >= 0:
            exp_hash = pow(10, self._exp, modulus)
        else:
            exp_hash = pow(pow(10, modulus - 2, modulus), -self._exp, modulus)
        hash_ = int(self._int) * exp_hash % modulus
        if self >= 0:
            ans = hash_
        else:
            ans = -hash_
        if ans == -1:
            return -2
        return ans

    def as_tuple(self):
        return DecimalTuple(self._sign, tuple(int(d) for d in self._int), self._exp)

    def as_integer_ratio(self):
        if self._is_special:
            if self.is_nan():
                raise ValueError("cannot convert NaN to integer ratio")
            else:
                raise OverflowError("cannot convert Infinity to integer ratio")
        if not self:
            return 0, 1
        n = int(self._int)
        if self._exp >= 0:
            n, d = n * 10 ** self._exp, 1
        else:
            d5 = -self._exp
            while d5 > 0 and n % 5 == 0:
                n //= 5
                d5 -= 1
            d2 = -self._exp
            shift2 = min((n & -n).bit_length() - 1, d2)
            if shift2:
                n >>= shift2
                d2 -= shift2
            d = 5 ** d5 << d2
        if self._sign:
            n = -n
        return n, d

    def __repr__(self):
        return "Decimal('" + str(self) + "')"

    def __str__(self):
        return self._str(False, None)

    def _str(self, eng=False, context=None):
        sign = ["", "-"][self._sign]
        if self._is_special:
            if self._exp == "F":
                return sign + "Infinity"
            elif self._exp == "n":
                return sign + "NaN" + self._int
            else:
                return sign + "sNaN" + self._int
        leftdigits = self._exp + len(self._int)
        if self._exp <= 0 and leftdigits > -6:
            dotplace = leftdigits
        elif not eng:
            dotplace = 1
        elif self._int == "0":
            dotplace = (leftdigits + 1) % 3 - 1
        else:
            dotplace = (leftdigits - 1) % 3 + 1
        if dotplace <= 0:
            intpart = "0"
            fracpart = "." + "0" * (-dotplace) + self._int
        elif dotplace >= len(self._int):
            intpart = self._int + "0" * (dotplace - len(self._int))
            fracpart = ""
        else:
            intpart = self._int[:dotplace]
            fracpart = "." + self._int[dotplace:]
        if leftdigits == dotplace:
            exp = ""
        else:
            if context is None:
                context = getcontext()
            exp = ["e", "E"][context.capitals] + "%+d" % (leftdigits - dotplace)
        return sign + intpart + fracpart + exp

    def to_eng_string(self, context=None):
        return self._str(True, context)

    def copy_negate_ctx(self, context=None):
        if self._is_special:
            ans = self._check_nans(None, context)
            if ans:
                return ans
        context = _getc(context)
        if not self and context.rounding != ROUND_FLOOR:
            ans = self.copy_abs()
        else:
            ans = self.copy_negate()
        return ans._fix(context)

    def _pos(self, context=None):
        if self._is_special:
            ans = self._check_nans(None, context)
            if ans:
                return ans
        if context is None:
            context = getcontext()
        if not self and context.rounding != ROUND_FLOOR:
            ans = self.copy_abs()
        else:
            ans = Decimal(self)
        return ans._fix(context)

    def _abs(self, round=True, context=None):
        if not round:
            return self.copy_abs()
        if self._is_special:
            ans = self._check_nans(None, context)
            if ans:
                return ans
        if self._sign:
            ans = self.copy_negate_ctx(context)
        else:
            ans = self._pos(context)
        return ans

    def _add(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        if context is None:
            context = getcontext()
        if self._is_special or other._is_special:
            ans = self._check_nans(other, context)
            if ans:
                return ans
            if self._isinfinity():
                if self._sign != other._sign and other._isinfinity():
                    return context._raise_error(InvalidOperation, "-INF + INF")
                return Decimal(self)
            if other._isinfinity():
                return Decimal(other)
        exp = min(self._exp, other._exp)
        negativezero = 0
        if context.rounding == ROUND_FLOOR and self._sign != other._sign:
            negativezero = 1
        if not self and not other:
            sign = min(self._sign, other._sign)
            if negativezero:
                sign = 1
            ans = _dec_from_triple(sign, "0", exp)
            ans = ans._fix(context)
            return ans
        if not self:
            exp = max(exp, other._exp - context.prec - 1)
            ans = other._rescale(exp, context.rounding)
            ans = ans._fix(context)
            return ans
        if not other:
            exp = max(exp, self._exp - context.prec - 1)
            ans = self._rescale(exp, context.rounding)
            ans = ans._fix(context)
            return ans
        op1 = _WorkRep(self)
        op2 = _WorkRep(other)
        op1, op2 = _normalize(op1, op2, context.prec)
        result = _WorkRep()
        if op1.sign != op2.sign:
            if op1.int == op2.int:
                ans = _dec_from_triple(negativezero, "0", exp)
                ans = ans._fix(context)
                return ans
            if op1.int < op2.int:
                op1, op2 = op2, op1
            if op1.sign == 1:
                result.sign = 1
                op1.sign, op2.sign = op2.sign, op1.sign
            else:
                result.sign = 0
        elif op1.sign == 1:
            result.sign = 1
            op1.sign, op2.sign = (0, 0)
        else:
            result.sign = 0
        if op2.sign == 0:
            result.int = op1.int + op2.int
        else:
            result.int = op1.int - op2.int
        result.exp = op1.exp
        ans = _dec_from_triple(result.sign, str(result.int), result.exp)
        ans = ans._fix(context)
        return ans

    def _sub(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        if self._is_special or other._is_special:
            ans = self._check_nans(other, context)
            if ans:
                return ans
        return self._add(other.copy_negate(), context=context)

    def _rsub(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        return other._sub(self, context=context)

    def _mul(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        if context is None:
            context = getcontext()
        resultsign = self._sign ^ other._sign
        if self._is_special or other._is_special:
            ans = self._check_nans(other, context)
            if ans:
                return ans
            if self._isinfinity():
                if not other:
                    return context._raise_error(InvalidOperation, "(+-)INF * 0")
                return _SignedInfinity[resultsign]
            if other._isinfinity():
                if not self:
                    return context._raise_error(InvalidOperation, "0 * (+-)INF")
                return _SignedInfinity[resultsign]
        resultexp = self._exp + other._exp
        if not self or not other:
            ans = _dec_from_triple(resultsign, "0", resultexp)
            ans = ans._fix(context)
            return ans
        if self._int == "1":
            ans = _dec_from_triple(resultsign, other._int, resultexp)
            ans = ans._fix(context)
            return ans
        if other._int == "1":
            ans = _dec_from_triple(resultsign, self._int, resultexp)
            ans = ans._fix(context)
            return ans
        op1 = int(self._int)
        op2 = int(other._int)
        ans = _dec_from_triple(resultsign, str(op1 * op2), resultexp)
        ans = ans._fix(context)
        return ans

    def _truediv(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return NotImplemented
        if context is None:
            context = getcontext()
        sign = self._sign ^ other._sign
        if self._is_special or other._is_special:
            ans = self._check_nans(other, context)
            if ans:
                return ans
            if self._isinfinity() and other._isinfinity():
                return context._raise_error(InvalidOperation, "(+-)INF/(+-)INF")
            if self._isinfinity():
                return _SignedInfinity[sign]
            if other._isinfinity():
                context._raise_error(Clamped, "Division by infinity")
                return _dec_from_triple(sign, "0", context.Etiny())
        if not other:
            if not self:
                return context._raise_error(DivisionUndefined, "0 / 0")
            return context._raise_error(DivisionByZero, "x / 0", sign)
        if not self:
            exp = self._exp - other._exp
            coeff = 0
        else:
            shift = len(other._int) - len(self._int) + context.prec + 1
            exp = self._exp - other._exp - shift
            op1 = _WorkRep(self)
            op2 = _WorkRep(other)
            if shift >= 0:
                coeff, remainder = divmod(op1.int * 10 ** shift, op2.int)
            else:
                coeff, remainder = divmod(op1.int, op2.int * 10 ** -shift)
            if remainder:
                if coeff % 5 == 0:
                    coeff += 1
            else:
                ideal_exp = self._exp - other._exp
                while exp < ideal_exp and coeff % 10 == 0:
                    coeff //= 10
                    exp += 1
        ans = _dec_from_triple(sign, str(coeff), exp)
        return ans._fix(context)

    def _rtruediv(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        return other._truediv(self, context=context)

    def _divide(self, other, context):
        sign = self._sign ^ other._sign
        if other._isinfinity():
            ideal_exp = self._exp
        else:
            ideal_exp = min(self._exp, other._exp)
        expdiff = self.adjusted() - other.adjusted()
        if not self or other._isinfinity() or expdiff <= -2:
            return (_dec_from_triple(sign, "0", 0), self._rescale(ideal_exp, context.rounding))
        if expdiff <= context.prec:
            op1 = _WorkRep(self)
            op2 = _WorkRep(other)
            if op1.exp >= op2.exp:
                op1.int *= 10 ** (op1.exp - op2.exp)
            else:
                op2.int *= 10 ** (op2.exp - op1.exp)
            q, r = divmod(op1.int, op2.int)
            if q < 10 ** context.prec:
                return (_dec_from_triple(sign, str(q), 0), _dec_from_triple(self._sign, str(r), ideal_exp))
        ans = context._raise_error(DivisionImpossible, "quotient too large in //, % or divmod")
        return ans, ans

    def _divmod(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        if context is None:
            context = getcontext()
        ans = self._check_nans(other, context)
        if ans:
            return (ans, ans)
        sign = self._sign ^ other._sign
        if self._isinfinity():
            if other._isinfinity():
                ans = context._raise_error(InvalidOperation, "divmod(INF, INF)")
                return ans, ans
            else:
                return (_SignedInfinity[sign], context._raise_error(InvalidOperation, "INF % x"))
        if not other:
            if not self:
                ans = context._raise_error(DivisionUndefined, "divmod(0, 0)")
                return ans, ans
            else:
                if context.traps[InvalidOperation]:
                    context.flags[InvalidOperation] = True
                    context.flags[DivisionByZero] = True
                    raise InvalidOperation("[<class 'decimal.InvalidOperation'>, <class 'decimal.DivisionByZero'>]")
                return (context._raise_error(DivisionByZero, "x // 0", sign), context._raise_error(InvalidOperation, "x % 0"))
        quotient, remainder = self._divide(other, context)
        if not other._isinfinity():
            quotient = quotient._fix(context)
        remainder = remainder._fix(context)
        return quotient, remainder

    def _rdivmod(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        return other._divmod(self, context=context)

    def _mod(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        if context is None:
            context = getcontext()
        ans = self._check_nans(other, context)
        if ans:
            return ans
        if self._isinfinity():
            return context._raise_error(InvalidOperation, "INF % x")
        elif not other:
            if self:
                return context._raise_error(InvalidOperation, "x % 0")
            else:
                return context._raise_error(DivisionUndefined, "0 % 0")
        remainder = self._divide(other, context)[1]
        remainder = remainder._fix(context)
        return remainder

    def _rmod(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        return other._mod(self, context=context)

    def remainder_near(self, other, context=None):
        if context is None:
            context = getcontext()
        other = _convert_other(other, raiseit=True)
        ans = self._check_nans(other, context)
        if ans:
            return ans
        if self._isinfinity():
            return context._raise_error(InvalidOperation, "remainder_near(infinity, x)")
        if not other:
            if self:
                return context._raise_error(InvalidOperation, "remainder_near(x, 0)")
            else:
                return context._raise_error(DivisionUndefined, "remainder_near(0, 0)")
        if other._isinfinity():
            ans = Decimal(self)
            return ans._fix(context)
        ideal_exponent = min(self._exp, other._exp)
        if not self:
            ans = _dec_from_triple(self._sign, "0", ideal_exponent)
            return ans._fix(context)
        expdiff = self.adjusted() - other.adjusted()
        if expdiff >= context.prec + 1:
            return context._raise_error(DivisionImpossible)
        if expdiff <= -2:
            ans = self._rescale(ideal_exponent, context.rounding)
            return ans._fix(context)
        op1 = _WorkRep(self)
        op2 = _WorkRep(other)
        if op1.exp >= op2.exp:
            op1.int *= 10 ** (op1.exp - op2.exp)
        else:
            op2.int *= 10 ** (op2.exp - op1.exp)
        q, r = divmod(op1.int, op2.int)
        if 2 * r + (q & 1) > op2.int:
            r -= op2.int
            q += 1
        if q >= 10 ** context.prec:
            return context._raise_error(DivisionImpossible)
        sign = self._sign
        if r < 0:
            sign = 1 - sign
            r = -r
        ans = _dec_from_triple(sign, str(r), ideal_exponent)
        return ans._fix(context)

    def _floordiv(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        if context is None:
            context = getcontext()
        ans = self._check_nans(other, context)
        if ans:
            return ans
        if self._isinfinity():
            if other._isinfinity():
                return context._raise_error(InvalidOperation, "INF // INF")
            else:
                return _SignedInfinity[self._sign ^ other._sign]
        if not other:
            if self:
                return context._raise_error(DivisionByZero, "x // 0", self._sign ^ other._sign)
            else:
                return context._raise_error(DivisionUndefined, "0 // 0")
        q = self._divide(other, context)[0]
        if other._isinfinity():
            return q
        return q._fix(context)

    def _rfloordiv(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        return other._floordiv(self, context=context)

    def __float__(self):
        if self._isnan():
            if self.is_snan():
                raise ValueError("cannot convert signaling NaN to float")
            s = "-nan" if self._sign else "nan"
        else:
            s = str(self)
        return float(s)

    def __int__(self):
        if self._is_special:
            if self._isnan():
                raise ValueError("cannot convert NaN to integer")
            elif self._isinfinity():
                raise OverflowError("cannot convert Infinity to integer")
        s = (-1) ** self._sign
        if self._exp >= 0:
            return s * int(self._int) * 10 ** self._exp
        else:
            return s * int(self._int[:self._exp] or "0")

    def __trunc__(self):
        return int(self)

    def real(self):
        return self

    def imag(self):
        return Decimal(0)

    def conjugate(self):
        return self

    def __complex__(self):
        return complex(float(self))

    def _fix_nan(self, context):
        payload = self._int
        max_payload_len = context.prec - context.clamp
        if len(payload) > max_payload_len:
            payload = payload[len(payload) - max_payload_len:].lstrip("0")
            return _dec_from_triple(self._sign, payload, self._exp, True)
        return Decimal(self)

    def _fix(self, context):
        if self._is_special:
            if self._isnan():
                return self._fix_nan(context)
            else:
                return Decimal(self)
        Etiny = context.Etiny()
        Etop = context.Etop()
        if not self:
            exp_max = [context.Emax, Etop][context.clamp]
            new_exp = min(max(self._exp, Etiny), exp_max)
            if new_exp != self._exp:
                context._raise_error(Clamped)
                return _dec_from_triple(self._sign, "0", new_exp)
            else:
                return Decimal(self)
        exp_min = len(self._int) + self._exp - context.prec
        if exp_min > Etop:
            ans = context._raise_error(Overflow, "above Emax", self._sign)
            context._raise_error(Inexact)
            context._raise_error(Rounded)
            return ans
        self_is_subnormal = exp_min < Etiny
        if self_is_subnormal:
            exp_min = Etiny
        if self._exp < exp_min:
            digits = len(self._int) + self._exp - exp_min
            if digits < 0:
                self = _dec_from_triple(self._sign, "1", exp_min - 1)
                digits = 0
            changed = self._round_by(context.rounding, digits)
            coeff = self._int[:digits] or "0"
            if changed > 0:
                coeff = str(int(coeff) + 1)
                if len(coeff) > context.prec:
                    coeff = coeff[:-1]
                    exp_min += 1
            if exp_min > Etop:
                ans = context._raise_error(Overflow, "above Emax", self._sign)
            else:
                ans = _dec_from_triple(self._sign, coeff, exp_min)
            if changed and self_is_subnormal:
                context._raise_error(Underflow)
            if self_is_subnormal:
                context._raise_error(Subnormal)
            if changed:
                context._raise_error(Inexact)
            context._raise_error(Rounded)
            if not ans:
                context._raise_error(Clamped)
            return ans
        if self_is_subnormal:
            context._raise_error(Subnormal)
        if context.clamp == 1 and self._exp > Etop:
            context._raise_error(Clamped)
            self_padded = self._int + "0" * (self._exp - Etop)
            return _dec_from_triple(self._sign, self_padded, Etop)
        return Decimal(self)

    def _round_down(self, prec):
        if _all_zeros(self._int, prec):
            return 0
        else:
            return -1

    def _round_up(self, prec):
        return -self._round_down(prec)

    def _round_half_up(self, prec):
        if self._int[prec] in "56789":
            return 1
        elif _all_zeros(self._int, prec):
            return 0
        else:
            return -1

    def _round_half_down(self, prec):
        if _exact_half(self._int, prec):
            return -1
        else:
            return self._round_half_up(prec)

    def _round_half_even(self, prec):
        if _exact_half(self._int, prec) and (prec == 0 or self._int[prec - 1] in "02468"):
            return -1
        else:
            return self._round_half_up(prec)

    def _round_ceiling(self, prec):
        if self._sign:
            return self._round_down(prec)
        else:
            return -self._round_down(prec)

    def _round_floor(self, prec):
        if not self._sign:
            return self._round_down(prec)
        else:
            return -self._round_down(prec)

    def _round_05up(self, prec):
        if prec and self._int[prec - 1] not in "05":
            return self._round_down(prec)
        else:
            return -self._round_down(prec)

    def _round_by(self, rounding, prec):
        if rounding == ROUND_DOWN:
            return self._round_down(prec)
        if rounding == ROUND_UP:
            return self._round_up(prec)
        if rounding == ROUND_HALF_UP:
            return self._round_half_up(prec)
        if rounding == ROUND_HALF_DOWN:
            return self._round_half_down(prec)
        if rounding == ROUND_HALF_EVEN:
            return self._round_half_even(prec)
        if rounding == ROUND_CEILING:
            return self._round_ceiling(prec)
        if rounding == ROUND_FLOOR:
            return self._round_floor(prec)
        return self._round_05up(prec)

    def _rescale(self, exp, rounding):
        if self._is_special:
            return Decimal(self)
        if not self:
            return _dec_from_triple(self._sign, "0", exp)
        if self._exp >= exp:
            return _dec_from_triple(self._sign, self._int + "0" * (self._exp - exp), exp)
        digits = len(self._int) + self._exp - exp
        if digits < 0:
            self = _dec_from_triple(self._sign, "1", exp - 1)
            digits = 0
        changed = self._round_by(rounding, digits)
        coeff = self._int[:digits] or "0"
        if changed == 1:
            coeff = str(int(coeff) + 1)
        return _dec_from_triple(self._sign, coeff, exp)

    def _round(self, places, rounding):
        if places <= 0:
            raise ValueError("argument should be at least 1 in _round")
        if self._is_special or not self:
            return Decimal(self)
        ans = self._rescale(self.adjusted() + 1 - places, rounding)
        if ans.adjusted() != self.adjusted():
            ans = ans._rescale(ans.adjusted() + 1 - places, rounding)
        return ans

    def to_integral_exact(self, rounding=None, context=None):
        if self._is_special:
            ans = self._check_nans(context=context)
            if ans:
                return ans
            return Decimal(self)
        if self._exp >= 0:
            return Decimal(self)
        if not self:
            return _dec_from_triple(self._sign, "0", 0)
        if context is None:
            context = getcontext()
        if rounding is None:
            rounding = context.rounding
        ans = self._rescale(0, rounding)
        if ans != self:
            context._raise_error(Inexact)
        context._raise_error(Rounded)
        return ans

    def to_integral_value(self, rounding=None, context=None):
        if context is None:
            context = getcontext()
        if rounding is None:
            rounding = context.rounding
        if self._is_special:
            ans = self._check_nans(context=context)
            if ans:
                return ans
            return Decimal(self)
        if self._exp >= 0:
            return Decimal(self)
        else:
            return self._rescale(0, rounding)

    def to_integral(self, rounding=None, context=None):
        return self.to_integral_value(rounding, context)

    def sqrt(self, context=None):
        if context is None:
            context = getcontext()
        if self._is_special:
            ans = self._check_nans(context=context)
            if ans:
                return ans
            if self._isinfinity() and self._sign == 0:
                return Decimal(self)
        if not self:
            ans = _dec_from_triple(self._sign, "0", self._exp // 2)
            return ans._fix(context)
        if self._sign == 1:
            return context._raise_error(InvalidOperation, "sqrt(-x), x > 0")
        prec = context.prec + 1
        op = _WorkRep(self)
        e = op.exp >> 1
        if op.exp & 1:
            c = op.int * 10
            l = (len(self._int) >> 1) + 1
        else:
            c = op.int
            l = len(self._int) + 1 >> 1
        shift = prec - l
        if shift >= 0:
            c *= 100 ** shift
            exact = True
        else:
            c, remainder = divmod(c, 100 ** -shift)
            exact = not remainder
        e -= shift
        n = 10 ** prec
        while True:
            q = c // n
            if n <= q:
                break
            else:
                n = n + q >> 1
        exact = exact and n * n == c
        if exact:
            if shift >= 0:
                n //= 10 ** shift
            else:
                n *= 10 ** -shift
            e += shift
        else:
            if n % 5 == 0:
                n += 1
        ans = _dec_from_triple(0, str(n), e)
        context = context._shallow_copy()
        rounding = context._set_rounding(ROUND_HALF_EVEN)
        ans = ans._fix(context)
        context.rounding = rounding
        return ans

    def max(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        if context is None:
            context = getcontext()
        if self._is_special or other._is_special:
            sn = self._isnan()
            on = other._isnan()
            if sn or on:
                if on == 1 and sn == 0:
                    return self._fix(context)
                if sn == 1 and on == 0:
                    return other._fix(context)
                return self._check_nans(other, context)
        c = self._cmp(other)
        if c == 0:
            c = self.compare_total(other)
        if c == -1:
            ans = other
        else:
            ans = self
        return ans._fix(context)

    def min(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        if context is None:
            context = getcontext()
        if self._is_special or other._is_special:
            sn = self._isnan()
            on = other._isnan()
            if sn or on:
                if on == 1 and sn == 0:
                    return self._fix(context)
                if sn == 1 and on == 0:
                    return other._fix(context)
                return self._check_nans(other, context)
        c = self._cmp(other)
        if c == 0:
            c = self.compare_total(other)
        if c == -1:
            ans = self
        else:
            ans = other
        return ans._fix(context)

    def _isinteger(self):
        if self._is_special:
            return False
        if self._exp >= 0:
            return True
        rest = self._int[self._exp:]
        return rest == "0" * len(rest)

    def _iseven(self):
        if not self or self._exp > 0:
            return True
        return self._int[-1 + self._exp] in "02468"

    def __add__(self, other):
        r = self._add(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __radd__(self, other):
        r = self._add(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __sub__(self, other):
        r = self._sub(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __rsub__(self, other):
        r = self._rsub(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __mul__(self, other):
        r = self._mul(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __rmul__(self, other):
        r = self._mul(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __truediv__(self, other):
        r = self._truediv(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __rtruediv__(self, other):
        r = self._rtruediv(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __divmod__(self, other):
        r = self._divmod(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __rdivmod__(self, other):
        r = self._rdivmod(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __mod__(self, other):
        r = self._mod(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __rmod__(self, other):
        r = self._rmod(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __floordiv__(self, other):
        r = self._floordiv(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __rfloordiv__(self, other):
        r = self._rfloordiv(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __pow__(self, other):
        r = self._pow(other, None, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __rpow__(self, other):
        r = self._rpow(other, None)
        if r is NotImplemented:
            return NotImplemented
        return r

    def __abs__(self):
        return self._abs(True, None)

    def __pos__(self):
        return self._pos()

    def __neg__(self):
        return self.copy_negate_ctx(None)

    def adjusted(self):
        try:
            return self._exp + len(self._int) - 1
        except TypeError:
            return 0

    def canonical(self):
        return self

    def compare_signal(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        ans = self._compare_check_nans(other, context)
        if ans:
            return ans
        return self.compare(other, context=context)

    def compare_total(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        if self._sign and not other._sign:
            return Decimal(-1)
        if not self._sign and other._sign:
            return Decimal(1)
        sign = self._sign
        self_nan = self._isnan()
        other_nan = other._isnan()
        if self_nan or other_nan:
            if self_nan == other_nan:
                self_key = len(self._int), self._int
                other_key = len(other._int), other._int
                if self_key < other_key:
                    if sign:
                        return Decimal(1)
                    else:
                        return Decimal(-1)
                if self_key > other_key:
                    if sign:
                        return Decimal(-1)
                    else:
                        return Decimal(1)
                return Decimal(0)
            if sign:
                if self_nan == 1:
                    return Decimal(-1)
                if other_nan == 1:
                    return Decimal(1)
                if self_nan == 2:
                    return Decimal(-1)
                if other_nan == 2:
                    return Decimal(1)
            else:
                if self_nan == 1:
                    return Decimal(1)
                if other_nan == 1:
                    return Decimal(-1)
                if self_nan == 2:
                    return Decimal(1)
                if other_nan == 2:
                    return Decimal(-1)
        if self < other:
            return Decimal(-1)
        if self > other:
            return Decimal(1)
        if self._exp < other._exp:
            if sign:
                return Decimal(1)
            else:
                return Decimal(-1)
        if self._exp > other._exp:
            if sign:
                return Decimal(-1)
            else:
                return Decimal(1)
        return Decimal(0)

    def compare_total_mag(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        s = self.copy_abs()
        o = other.copy_abs()
        return s.compare_total(o)

    def copy_abs(self):
        return _dec_from_triple(0, self._int, self._exp, self._is_special)

    def copy_negate(self):
        if self._sign:
            return _dec_from_triple(0, self._int, self._exp, self._is_special)
        else:
            return _dec_from_triple(1, self._int, self._exp, self._is_special)

    def copy_sign(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        return _dec_from_triple(other._sign, self._int, self._exp, self._is_special)

    def exp(self, context=None):
        if context is None:
            context = getcontext()
        ans = self._check_nans(context=context)
        if ans:
            return ans
        if self._isinfinity() == -1:
            return _Zero
        if not self:
            return _One
        if self._isinfinity() == 1:
            return Decimal(self)
        p = context.prec
        adj = self.adjusted()
        if self._sign == 0 and adj > len(str((context.Emax + 1) * 3)):
            ans = _dec_from_triple(0, "1", context.Emax + 1)
        elif self._sign == 1 and adj > len(str((-context.Etiny() + 1) * 3)):
            ans = _dec_from_triple(0, "1", context.Etiny() - 1)
        elif self._sign == 0 and adj < -p:
            ans = _dec_from_triple(0, "1" + "0" * (p - 1) + "1", -p)
        elif self._sign == 1 and adj < -p - 1:
            ans = _dec_from_triple(0, "9" * (p + 1), -p - 1)
        else:
            op = _WorkRep(self)
            c, e = op.int, op.exp
            if op.sign == 1:
                c = -c
            extra = 3
            while True:
                coeff, exp = _dexp(c, e, p + extra)
                if coeff % (5 * 10 ** (len(str(coeff)) - p - 1)):
                    break
                extra += 3
            ans = _dec_from_triple(0, str(coeff), exp)
        context = context._shallow_copy()
        rounding = context._set_rounding(ROUND_HALF_EVEN)
        ans = ans._fix(context)
        context.rounding = rounding
        return ans

    def is_canonical(self):
        return True

    def is_finite(self):
        return not self._is_special

    def is_infinite(self):
        return self._exp == "F"

    def is_nan(self):
        return self._exp in ("n", "N")

    def is_normal(self, context=None):
        if self._is_special or not self:
            return False
        if context is None:
            context = getcontext()
        return context.Emin <= self.adjusted()

    def is_qnan(self):
        return self._exp == "n"

    def is_signed(self):
        return self._sign == 1

    def is_snan(self):
        return self._exp == "N"

    def is_subnormal(self, context=None):
        if self._is_special or not self:
            return False
        if context is None:
            context = getcontext()
        return self.adjusted() < context.Emin

    def is_zero(self):
        return not self._is_special and self._int == "0"

    def _ln_exp_bound(self):
        adj = self._exp + len(self._int) - 1
        if adj >= 1:
            return len(str(adj * 23 // 10)) - 1
        if adj <= -2:
            return len(str((-1 - adj) * 23 // 10)) - 1
        op = _WorkRep(self)
        c, e = op.int, op.exp
        if adj == 0:
            num = str(c - 10 ** -e)
            den = str(c)
            return len(num) - len(den) - (num < den)
        return e + len(str(10 ** -e - c)) - 1

    def ln(self, context=None):
        if context is None:
            context = getcontext()
        ans = self._check_nans(context=context)
        if ans:
            return ans
        if not self:
            return _NegativeInfinity
        if self._isinfinity() == 1:
            return _Infinity
        if self == _One:
            return _Zero
        if self._sign == 1:
            return context._raise_error(InvalidOperation, "ln of a negative value")
        op = _WorkRep(self)
        c, e = op.int, op.exp
        p = context.prec
        places = p - self._ln_exp_bound() + 2
        while True:
            coeff = _dlog(c, e, places)
            if coeff % (5 * 10 ** (len(str(abs(coeff))) - p - 1)):
                break
            places += 3
        ans = _dec_from_triple(int(coeff < 0), str(abs(coeff)), -places)
        context = context._shallow_copy()
        rounding = context._set_rounding(ROUND_HALF_EVEN)
        ans = ans._fix(context)
        context.rounding = rounding
        return ans

    def _log10_exp_bound(self):
        adj = self._exp + len(self._int) - 1
        if adj >= 1:
            return len(str(adj)) - 1
        if adj <= -2:
            return len(str(-1 - adj)) - 1
        op = _WorkRep(self)
        c, e = op.int, op.exp
        if adj == 0:
            num = str(c - 10 ** -e)
            den = str(231 * c)
            return len(num) - len(den) - (num < den) + 2
        num = str(10 ** -e - c)
        return len(num) + e - (num < "231") - 1

    def log10(self, context=None):
        if context is None:
            context = getcontext()
        ans = self._check_nans(context=context)
        if ans:
            return ans
        if not self:
            return _NegativeInfinity
        if self._isinfinity() == 1:
            return _Infinity
        if self._sign == 1:
            return context._raise_error(InvalidOperation, "log10 of a negative value")
        if self._int[0] == "1" and self._int[1:] == "0" * (len(self._int) - 1):
            ans = Decimal(self._exp + len(self._int) - 1)
        else:
            op = _WorkRep(self)
            c, e = op.int, op.exp
            p = context.prec
            places = p - self._log10_exp_bound() + 2
            while True:
                coeff = _dlog10(c, e, places)
                if coeff % (5 * 10 ** (len(str(abs(coeff))) - p - 1)):
                    break
                places += 3
            ans = _dec_from_triple(int(coeff < 0), str(abs(coeff)), -places)
        context = context._shallow_copy()
        rounding = context._set_rounding(ROUND_HALF_EVEN)
        ans = ans._fix(context)
        context.rounding = rounding
        return ans

    def logb(self, context=None):
        ans = self._check_nans(context=context)
        if ans:
            return ans
        if context is None:
            context = getcontext()
        if self._isinfinity():
            return _Infinity
        if not self:
            return context._raise_error(DivisionByZero, "logb(0)", 1)
        ans = Decimal(self.adjusted())
        return ans._fix(context)

    def _islogical(self):
        if self._sign != 0 or self._exp != 0:
            return False
        for dig in self._int:
            if dig not in "01":
                return False
        return True

    def _fill_logical(self, context, opa, opb):
        dif = context.prec - len(opa)
        if dif > 0:
            opa = "0" * dif + opa
        elif dif < 0:
            opa = opa[-context.prec:]
        dif = context.prec - len(opb)
        if dif > 0:
            opb = "0" * dif + opb
        elif dif < 0:
            opb = opb[-context.prec:]
        return opa, opb

    def logical_and(self, other, context=None):
        if context is None:
            context = getcontext()
        other = _convert_other(other, raiseit=True)
        if not self._islogical() or not other._islogical():
            return context._raise_error(InvalidOperation)
        opa, opb = self._fill_logical(context, self._int, other._int)
        result = "".join([str(int(a) & int(b)) for a, b in zip(opa, opb)])
        return _dec_from_triple(0, result.lstrip("0") or "0", 0)

    def logical_invert(self, context=None):
        if context is None:
            context = getcontext()
        return self.logical_xor(_dec_from_triple(0, "1" * context.prec, 0), context)

    def logical_or(self, other, context=None):
        if context is None:
            context = getcontext()
        other = _convert_other(other, raiseit=True)
        if not self._islogical() or not other._islogical():
            return context._raise_error(InvalidOperation)
        opa, opb = self._fill_logical(context, self._int, other._int)
        result = "".join([str(int(a) | int(b)) for a, b in zip(opa, opb)])
        return _dec_from_triple(0, result.lstrip("0") or "0", 0)

    def logical_xor(self, other, context=None):
        if context is None:
            context = getcontext()
        other = _convert_other(other, raiseit=True)
        if not self._islogical() or not other._islogical():
            return context._raise_error(InvalidOperation)
        opa, opb = self._fill_logical(context, self._int, other._int)
        result = "".join([str(int(a) ^ int(b)) for a, b in zip(opa, opb)])
        return _dec_from_triple(0, result.lstrip("0") or "0", 0)

    def max_mag(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        if context is None:
            context = getcontext()
        if self._is_special or other._is_special:
            sn = self._isnan()
            on = other._isnan()
            if sn or on:
                if on == 1 and sn == 0:
                    return self._fix(context)
                if sn == 1 and on == 0:
                    return other._fix(context)
                return self._check_nans(other, context)
        c = self.copy_abs()._cmp(other.copy_abs())
        if c == 0:
            c = self.compare_total(other)
        if c == -1:
            ans = other
        else:
            ans = self
        return ans._fix(context)

    def min_mag(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        if context is None:
            context = getcontext()
        if self._is_special or other._is_special:
            sn = self._isnan()
            on = other._isnan()
            if sn or on:
                if on == 1 and sn == 0:
                    return self._fix(context)
                if sn == 1 and on == 0:
                    return other._fix(context)
                return self._check_nans(other, context)
        c = self.copy_abs()._cmp(other.copy_abs())
        if c == 0:
            c = self.compare_total(other)
        if c == -1:
            ans = self
        else:
            ans = other
        return ans._fix(context)

    def next_minus(self, context=None):
        if context is None:
            context = getcontext()
        ans = self._check_nans(context=context)
        if ans:
            return ans
        if self._isinfinity() == -1:
            return _NegativeInfinity
        if self._isinfinity() == 1:
            return _dec_from_triple(0, "9" * context.prec, context.Etop())
        context = context.copy()
        context._set_rounding(ROUND_FLOOR)
        context._ignore_all_flags()
        new_self = self._fix(context)
        if new_self != self:
            return new_self
        return self._sub(_dec_from_triple(0, "1", context.Etiny() - 1), context)

    def next_plus(self, context=None):
        if context is None:
            context = getcontext()
        ans = self._check_nans(context=context)
        if ans:
            return ans
        if self._isinfinity() == 1:
            return _Infinity
        if self._isinfinity() == -1:
            return _dec_from_triple(1, "9" * context.prec, context.Etop())
        context = context.copy()
        context._set_rounding(ROUND_CEILING)
        context._ignore_all_flags()
        new_self = self._fix(context)
        if new_self != self:
            return new_self
        return self._add(_dec_from_triple(0, "1", context.Etiny() - 1), context)

    def next_toward(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        if context is None:
            context = getcontext()
        ans = self._check_nans(other, context)
        if ans:
            return ans
        comparison = self._cmp(other)
        if comparison == 0:
            return self.copy_sign(other)
        if comparison == -1:
            ans = self.next_plus(context)
        else:
            ans = self.next_minus(context)
        if ans._isinfinity():
            context._raise_error(Overflow, "Infinite result from next_toward", ans._sign)
            context._raise_error(Inexact)
            context._raise_error(Rounded)
        elif ans.adjusted() < context.Emin:
            context._raise_error(Underflow)
            context._raise_error(Subnormal)
            context._raise_error(Inexact)
            context._raise_error(Rounded)
            if not ans:
                context._raise_error(Clamped)
        return ans

    def number_class(self, context=None):
        if self.is_snan():
            return "sNaN"
        if self.is_qnan():
            return "NaN"
        inf = self._isinfinity()
        if inf == 1:
            return "+Infinity"
        if inf == -1:
            return "-Infinity"
        if self.is_zero():
            if self._sign:
                return "-Zero"
            else:
                return "+Zero"
        if context is None:
            context = getcontext()
        if self.is_subnormal(context=context):
            if self._sign:
                return "-Subnormal"
            else:
                return "+Subnormal"
        if self._sign:
            return "-Normal"
        else:
            return "+Normal"

    def radix(self):
        return Decimal(10)

    def rotate(self, other, context=None):
        if context is None:
            context = getcontext()
        other = _convert_other(other, raiseit=True)
        ans = self._check_nans(other, context)
        if ans:
            return ans
        if other._exp != 0:
            return context._raise_error(InvalidOperation)
        if not (-context.prec <= int(other) <= context.prec):
            return context._raise_error(InvalidOperation)
        if self._isinfinity():
            return Decimal(self)
        torot = int(other)
        rotdig = self._int
        topad = context.prec - len(rotdig)
        if topad > 0:
            rotdig = "0" * topad + rotdig
        elif topad < 0:
            rotdig = rotdig[-topad:]
        rotated = rotdig[torot:] + rotdig[:torot]
        return _dec_from_triple(self._sign, rotated.lstrip("0") or "0", self._exp)._fix_ctx(context)

    def _fix_ctx(self, context):
        return self._fix(context)

    def scaleb(self, other, context=None):
        if context is None:
            context = getcontext()
        other = _convert_other(other, raiseit=True)
        ans = self._check_nans(other, context)
        if ans:
            return ans
        if other._exp != 0:
            return context._raise_error(InvalidOperation)
        liminf = -2 * (context.Emax + context.prec)
        limsup = 2 * (context.Emax + context.prec)
        if not (liminf <= int(other) <= limsup):
            return context._raise_error(InvalidOperation)
        if self._isinfinity():
            return Decimal(self)
        d = _dec_from_triple(self._sign, self._int, self._exp + int(other))
        d = d._fix(context)
        return d

    def shift(self, other, context=None):
        if context is None:
            context = getcontext()
        other = _convert_other(other, raiseit=True)
        ans = self._check_nans(other, context)
        if ans:
            return ans
        if other._exp != 0:
            return context._raise_error(InvalidOperation)
        if not (-context.prec <= int(other) <= context.prec):
            return context._raise_error(InvalidOperation)
        if self._isinfinity():
            return Decimal(self)
        torot = int(other)
        rotdig = self._int
        topad = context.prec - len(rotdig)
        if topad > 0:
            rotdig = "0" * topad + rotdig
        elif topad < 0:
            rotdig = rotdig[-topad:]
        if torot < 0:
            shifted = rotdig[:torot]
        else:
            shifted = rotdig + "0" * torot
            shifted = shifted[-context.prec:]
        return _dec_from_triple(self._sign, shifted.lstrip("0") or "0", self._exp)._fix_ctx(context)

    def quantize(self, exp, rounding=None, context=None):
        exp = _convert_other(exp, raiseit=True)
        if context is None:
            context = getcontext()
        if rounding is None:
            rounding = context.rounding
        if self._is_special or exp._is_special:
            ans = self._check_nans(exp, context)
            if ans:
                return ans
            if exp._isinfinity() or self._isinfinity():
                if exp._isinfinity() and self._isinfinity():
                    return Decimal(self)
                return context._raise_error(InvalidOperation, "quantize with one INF")
        if not (context.Etiny() <= exp._exp <= context.Emax):
            return context._raise_error(InvalidOperation, "target exponent out of bounds in quantize")
        if not self:
            ans = _dec_from_triple(self._sign, "0", exp._exp)
            return ans._fix(context)
        self_adjusted = self.adjusted()
        if self_adjusted > context.Emax:
            return context._raise_error(InvalidOperation, "exponent of quantize result too large for current context")
        if self_adjusted - exp._exp + 1 > context.prec:
            return context._raise_error(InvalidOperation, "quantize result has too many digits for current context")
        ans = self._rescale(exp._exp, rounding)
        if ans.adjusted() > context.Emax:
            return context._raise_error(InvalidOperation, "exponent of quantize result too large for current context")
        if len(ans._int) > context.prec:
            return context._raise_error(InvalidOperation, "quantize result has too many digits for current context")
        if ans and ans.adjusted() < context.Emin:
            context._raise_error(Subnormal)
        if ans._exp > self._exp:
            if ans != self:
                context._raise_error(Inexact)
            context._raise_error(Rounded)
        ans = ans._fix(context)
        return ans

    def same_quantum(self, other, context=None):
        other = _convert_other(other, raiseit=True)
        if self._is_special or other._is_special:
            return (self.is_nan() and other.is_nan() or self.is_infinite() and other.is_infinite())
        return self._exp == other._exp

    def normalize(self, context=None):
        if context is None:
            context = getcontext()
        if self._is_special:
            ans = self._check_nans(context=context)
            if ans:
                return ans
        dup = self._fix(context)
        if dup._isinfinity():
            return dup
        if not dup:
            return _dec_from_triple(dup._sign, "0", 0)
        exp_max = [context.Emax, context.Etop()][context.clamp]
        end = len(dup._int)
        exp = dup._exp
        while dup._int[end - 1] == "0" and exp < exp_max:
            exp += 1
            end -= 1
        return _dec_from_triple(dup._sign, dup._int[:end], exp)

    def fma(self, other, third, context=None):
        other = _convert_other(other, raiseit=True)
        third = _convert_other(third, raiseit=True)
        if self._is_special or other._is_special:
            if context is None:
                context = getcontext()
            if self._exp == "N":
                return context._raise_error(InvalidOperation, "sNaN", self)
            if other._exp == "N":
                return context._raise_error(InvalidOperation, "sNaN", other)
            if self._exp == "n":
                product = self
            elif other._exp == "n":
                product = other
            elif self._exp == "F":
                if not other:
                    return context._raise_error(InvalidOperation, "INF * 0 in fma")
                product = _SignedInfinity[self._sign ^ other._sign]
            elif other._exp == "F":
                if not self:
                    return context._raise_error(InvalidOperation, "0 * INF in fma")
                product = _SignedInfinity[self._sign ^ other._sign]
        else:
            product = _dec_from_triple(self._sign ^ other._sign, str(int(self._int) * int(other._int)), self._exp + other._exp)
        return product._add(third, context)

    def _power_modulo(self, other, modulo, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        modulo = _convert_other(modulo)
        if modulo is NotImplemented:
            return modulo
        if context is None:
            context = getcontext()
        self_is_nan = self._isnan()
        other_is_nan = other._isnan()
        modulo_is_nan = modulo._isnan()
        if self_is_nan or other_is_nan or modulo_is_nan:
            if self_is_nan == 2:
                return context._raise_error(InvalidOperation, "sNaN", self)
            if other_is_nan == 2:
                return context._raise_error(InvalidOperation, "sNaN", other)
            if modulo_is_nan == 2:
                return context._raise_error(InvalidOperation, "sNaN", modulo)
            if self_is_nan:
                return self._fix_nan(context)
            if other_is_nan:
                return other._fix_nan(context)
            return modulo._fix_nan(context)
        if not (self._isinteger() and other._isinteger() and modulo._isinteger()):
            return context._raise_error(InvalidOperation, "pow() 3rd argument not allowed unless all arguments are integers")
        if other < 0:
            return context._raise_error(InvalidOperation, "pow() 2nd argument cannot be negative when 3rd argument specified")
        if not modulo:
            return context._raise_error(InvalidOperation, "pow() 3rd argument cannot be 0")
        if modulo.adjusted() >= context.prec:
            return context._raise_error(InvalidOperation, "insufficient precision: pow() 3rd argument must not have more than precision digits")
        if not other and not self:
            return context._raise_error(InvalidOperation, "at least one of pow() 1st argument and 2nd argument must be nonzero; 0**0 is not defined")
        if other._iseven():
            sign = 0
        else:
            sign = self._sign
        modulo = abs(int(modulo))
        base = _WorkRep(self.to_integral_value())
        exponent = _WorkRep(other.to_integral_value())
        base = (base.int % modulo * pow(10, base.exp, modulo)) % modulo
        for i in range(exponent.exp):
            base = pow(base, 10, modulo)
        base = pow(base, exponent.int, modulo)
        return _dec_from_triple(sign, str(base), 0)

    def _power_exact(self, other, p):
        x = _WorkRep(self)
        xc, xe = x.int, x.exp
        while xc % 10 == 0:
            xc //= 10
            xe += 1
        y = _WorkRep(other)
        yc, ye = y.int, y.exp
        while yc % 10 == 0:
            yc //= 10
            ye += 1
        if xc == 1:
            xe *= yc
            while xe % 10 == 0:
                xe //= 10
                ye += 1
            if ye < 0:
                return None
            exponent = xe * 10 ** ye
            if y.sign == 1:
                exponent = -exponent
            if other._isinteger() and other._sign == 0:
                ideal_exponent = self._exp * int(other)
                zeros = min(exponent - ideal_exponent, p - 1)
            else:
                zeros = 0
            return _dec_from_triple(0, "1" + "0" * zeros, exponent - zeros)
        if y.sign == 1:
            last_digit = xc % 10
            if last_digit in (2, 4, 6, 8):
                if xc & -xc != xc:
                    return None
                e = _nbits(xc) - 1
                emax = p * 93 // 65
                if ye >= len(str(emax)):
                    return None
                e = _decimal_lshift_exact(e * yc, ye)
                xe = _decimal_lshift_exact(xe * yc, ye)
                if e is None or xe is None:
                    return None
                if e > emax:
                    return None
                xc = 5 ** e
            elif last_digit == 5:
                e = _nbits(xc) * 28 // 65
                xc, remainder = divmod(5 ** e, xc)
                if remainder:
                    return None
                while xc % 5 == 0:
                    xc //= 5
                    e -= 1
                emax = p * 10 // 3
                if ye >= len(str(emax)):
                    return None
                e = _decimal_lshift_exact(e * yc, ye)
                xe = _decimal_lshift_exact(xe * yc, ye)
                if e is None or xe is None:
                    return None
                if e > emax:
                    return None
                xc = 2 ** e
            else:
                return None
            if xc >= 10 ** p:
                return None
            xe = -e - xe
            return _dec_from_triple(0, str(xc), xe)
        if ye >= 0:
            m, n = yc * 10 ** ye, 1
        else:
            if xe != 0 and len(str(abs(yc * xe))) <= -ye:
                return None
            xc_bits = _nbits(xc)
            if len(str(abs(yc) * xc_bits)) <= -ye:
                return None
            m, n = yc, 10 ** (-ye)
            while m % 2 == 0 and n % 2 == 0:
                m //= 2
                n //= 2
            while m % 5 == 0 and n % 5 == 0:
                m //= 5
                n //= 5
        if n > 1:
            if xc != 1 and xc_bits <= n:
                return None
            xe, rem = divmod(xe, n)
            if rem != 0:
                return None
            a = 1 << -(-_nbits(xc) // n)
            while True:
                q, r = divmod(xc, a ** (n - 1))
                if a <= q:
                    break
                else:
                    a = (a * (n - 1) + q) // n
            if not (a == q and r == 0):
                return None
            xc = a
        if xc > 1 and m > p * 100 // _log10_lb(xc):
            return None
        xc = xc ** m
        xe *= m
        if xc > 10 ** p:
            return None
        str_xc = str(xc)
        if other._isinteger() and other._sign == 0:
            ideal_exponent = self._exp * int(other)
            zeros = min(xe - ideal_exponent, p - len(str_xc))
        else:
            zeros = 0
        return _dec_from_triple(0, str_xc + "0" * zeros, xe - zeros)

    def _pow(self, other, modulo=None, context=None):
        if modulo is not None:
            return self._power_modulo(other, modulo, context)
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        if context is None:
            context = getcontext()
        ans = self._check_nans(other, context)
        if ans:
            return ans
        if not other:
            if not self:
                return context._raise_error(InvalidOperation, "0 ** 0")
            else:
                return _One
        result_sign = 0
        if self._sign == 1:
            if other._isinteger():
                if not other._iseven():
                    result_sign = 1
            else:
                if self:
                    return context._raise_error(InvalidOperation, "x ** y with x negative and y not an integer")
            self = self.copy_negate()
        if not self:
            if other._sign == 0:
                return _dec_from_triple(result_sign, "0", 0)
            else:
                return _SignedInfinity[result_sign]
        if self._isinfinity():
            if other._sign == 0:
                return _SignedInfinity[result_sign]
            else:
                return _dec_from_triple(result_sign, "0", 0)
        if self == _One:
            if other._isinteger():
                if other._sign == 1:
                    multiplier = 0
                elif other > context.prec:
                    multiplier = context.prec
                else:
                    multiplier = int(other)
                exp = self._exp * multiplier
                if exp < 1 - context.prec:
                    exp = 1 - context.prec
                    context._raise_error(Rounded)
            else:
                context._raise_error(Inexact)
                context._raise_error(Rounded)
                exp = 1 - context.prec
            return _dec_from_triple(result_sign, "1" + "0" * -exp, exp)
        self_adj = self.adjusted()
        if other._isinfinity():
            if (other._sign == 0) == (self_adj < 0):
                return _dec_from_triple(result_sign, "0", 0)
            else:
                return _SignedInfinity[result_sign]
        ans = None
        exact = False
        bound = self._log10_exp_bound() + other.adjusted()
        if (self_adj >= 0) == (other._sign == 0):
            if bound >= len(str(context.Emax)):
                ans = _dec_from_triple(result_sign, "1", context.Emax + 1)
        else:
            Etiny = context.Etiny()
            if bound >= len(str(-Etiny)):
                zero = _dec_from_triple(result_sign, "0", Etiny)
                for exc in (Underflow, Subnormal, Inexact, Rounded, Clamped):
                    context._raise_error(exc)
                return zero
        if ans is None:
            ans = self._power_exact(other, context.prec + 1)
            if ans is not None:
                if result_sign == 1:
                    ans = _dec_from_triple(1, ans._int, ans._exp)
                exact = True
        if ans is None:
            p = context.prec
            x = _WorkRep(self)
            xc, xe = x.int, x.exp
            y = _WorkRep(other)
            yc, ye = y.int, y.exp
            if y.sign == 1:
                yc = -yc
            extra = 3
            while True:
                coeff, exp = _dpower(xc, xe, yc, ye, p + extra)
                if coeff % (5 * 10 ** (len(str(coeff)) - p - 1)):
                    break
                extra += 3
            ans = _dec_from_triple(result_sign, str(coeff), exp)
        if ans.copy_abs()._cmp(_dec_from_triple(0, "5", context.Etiny() - 5)) < 0:
            zero = _dec_from_triple(ans._sign, "0", context.Etiny())
            for exc in (Underflow, Subnormal, Inexact, Rounded, Clamped):
                context._raise_error(exc)
            return zero
        if exact and not other._isinteger():
            if len(ans._int) <= context.prec:
                expdiff = context.prec + 1 - len(ans._int)
                ans = _dec_from_triple(ans._sign, ans._int + "0" * expdiff, ans._exp - expdiff)
            newcontext = context.copy()
            newcontext.clear_flags()
            for exception in _signals:
                newcontext.traps[exception] = False
            ans = ans._fix(newcontext)
            newcontext._raise_error(Inexact)
            if newcontext.flags[Subnormal]:
                newcontext._raise_error(Underflow)
            if newcontext.flags[Overflow]:
                context._raise_error(Overflow, "above Emax", ans._sign)
            for exception in (Underflow, Subnormal, Inexact, Rounded, Clamped):
                if newcontext.flags[exception]:
                    context._raise_error(exception)
        else:
            ans = ans._fix(context)
        return ans

    def _rpow(self, other, context=None):
        other = _convert_other(other)
        if other is NotImplemented:
            return other
        return other._pow(self, context=context)

    def __round__(self, n=None):
        if n is not None:
            if not isinstance(n, int):
                raise TypeError("Second argument to round should be integral")
            exp = _dec_from_triple(0, "1", -n)
            return self.quantize(exp)
        if self._is_special:
            if self.is_nan():
                raise ValueError("cannot convert NaN to integer")
            else:
                raise OverflowError("cannot convert Infinity to integer")
        return int(self._rescale(0, ROUND_HALF_EVEN))

    def __floor__(self):
        if self._is_special:
            if self.is_nan():
                raise ValueError("cannot convert NaN to integer")
            else:
                raise OverflowError("cannot convert Infinity to integer")
        return int(self._rescale(0, ROUND_FLOOR))

    def __ceil__(self):
        if self._is_special:
            if self.is_nan():
                raise ValueError("cannot convert NaN to integer")
            else:
                raise OverflowError("cannot convert Infinity to integer")
        return int(self._rescale(0, ROUND_CEILING))

    def __format__(self, specifier):
        return self._format(specifier, None)

    def _format(self, specifier, context=None):
        spec = _parse_format_specifier(specifier)
        if context is None:
            context = getcontext()
        if self._is_special:
            sign = _format_sign(self._sign, spec)
            body = str(self.copy_abs())
            if spec["type"] == "%":
                body += "%"
            return _format_align(sign, body, spec)
        if spec["type"] is None:
            spec["type"] = ["g", "G"][context.capitals]
        if spec["type"] == "%":
            self = _dec_from_triple(self._sign, self._int, self._exp + 2)
        rounding = context.rounding
        precision = spec["precision"]
        if precision is not None:
            if spec["type"] in "eE":
                self = self._round(precision + 1, rounding)
            elif spec["type"] in "fF%":
                self = self._rescale(-precision, rounding)
            elif spec["type"] in "gG" and len(self._int) > precision:
                self = self._round(precision, rounding)
        if not self and self._exp > 0 and spec["type"] in "fF%":
            self = self._rescale(0, rounding)
        if not self and spec["no_neg_0"] and self._sign:
            adjusted_sign = 0
        else:
            adjusted_sign = self._sign
        leftdigits = self._exp + len(self._int)
        if spec["type"] in "eE":
            if not self and precision is not None:
                dotplace = 1 - precision
            else:
                dotplace = 1
        elif spec["type"] in "fF%":
            dotplace = leftdigits
        else:
            if self._exp <= 0 and leftdigits > -6:
                dotplace = leftdigits
            else:
                dotplace = 1
        if dotplace < 0:
            intpart = "0"
            fracpart = "0" * (-dotplace) + self._int
        elif dotplace > len(self._int):
            intpart = self._int + "0" * (dotplace - len(self._int))
            fracpart = ""
        else:
            intpart = self._int[:dotplace] or "0"
            fracpart = self._int[dotplace:]
        exp = leftdigits - dotplace
        return _format_number(adjusted_sign, intpart, fracpart, exp, spec)


def _format_sign(is_negative, spec):
    if is_negative:
        return "-"
    elif spec["sign"] in " +":
        return spec["sign"]
    else:
        return ""


def _format_align(sign, body, spec):
    minimumwidth = spec["minimumwidth"]
    fill = spec["fill"]
    padding = fill * (minimumwidth - len(sign) - len(body))
    align = spec["align"]
    if align == "<":
        result = sign + body + padding
    elif align == ">":
        result = padding + sign + body
    elif align == "=":
        result = sign + padding + body
    elif align == "^":
        half = len(padding) // 2
        result = padding[:half] + sign + body + padding[half:]
    else:
        raise ValueError("Unrecognised alignment field")
    return result


def _group_lengths(grouping):
    out = []
    for g in grouping:
        out.append(g)
    return out


def _insert_thousands_sep(digits, spec, min_width=1):
    sep = spec["thousands_sep"]
    if not sep:
        return digits
    groups = []
    done = False
    while True:
        l = min(max(len(digits), min_width, 1), 3)
        groups.append("0" * (l - len(digits[-l:])) + digits[-l:])
        digits = digits[:-l]
        min_width -= l
        if not digits and min_width <= 0:
            done = True
            break
        min_width -= len(sep)
        if len(groups) > 1000:
            break
    groups.reverse()
    return sep.join(groups)


def _format_number(is_negative, intpart, fracpart, exp, spec):
    sign = _format_sign(is_negative, spec)
    if fracpart or spec["alt"]:
        fracpart = "." + fracpart
    if exp != 0 or spec["type"] in "eE":
        echar = {"E": "E", "e": "e", "G": "E", "g": "e"}[spec["type"]]
        fracpart += echar + "%+d" % exp
    if spec["type"] == "%":
        fracpart += "%"
    if spec["zeropad"]:
        min_width = spec["minimumwidth"] - len(fracpart) - len(sign)
    else:
        min_width = 0
    intpart = _insert_thousands_sep(intpart, spec, min_width)
    return _format_align(sign, intpart + fracpart, spec)


def _parse_format_specifier(specifier):
    m = _re.match(r"\A(?:(?P<fill>.)?(?P<align>[<>=^]))?(?P<sign>[-+ ])?(?P<no_neg_0>z)?(?P<alt>\#)?(?P<zeropad>0)?(?P<minimumwidth>(?!0)\d+)?(?P<thousands_sep>[,_])?(?:\.(?P<precision>\d+))?(?P<type>[eEfFgGn%])?\Z", specifier, _re.DOTALL)
    if m is None:
        raise ValueError("Invalid format specifier: " + specifier)
    fill = m.group("fill")
    align = m.group("align")
    out = {}
    zeropad = m.group("zeropad") is not None
    if zeropad:
        if fill is not None:
            raise ValueError("Fill character conflicts with '0' in format specifier: " + specifier)
        if align is not None:
            raise ValueError("Alignment conflicts with '0' in format specifier: " + specifier)
        out["fill"] = "0"
        out["align"] = "="
    else:
        out["fill"] = fill or " "
        out["align"] = align or ">"
    out["zeropad"] = zeropad
    out["sign"] = m.group("sign") or "-"
    out["no_neg_0"] = bool(m.group("no_neg_0"))
    out["alt"] = bool(m.group("alt"))
    mw = m.group("minimumwidth")
    out["minimumwidth"] = int(mw) if mw else 0
    out["thousands_sep"] = m.group("thousands_sep") or ""
    p = m.group("precision")
    out["precision"] = int(p) if p is not None else None
    t = m.group("type")
    if t == "n":
        t = "g"
    out["type"] = t
    if out["type"] in ("g", "G") and out["precision"] == 0:
        out["precision"] = 1
    return out


def _all_zeros(s, prec):
    for ch in s[prec:]:
        if ch != "0":
            return False
    return True


def _exact_half(s, prec):
    rest = s[prec:]
    if not rest:
        return False
    if rest[0] != "5":
        return False
    for ch in rest[1:]:
        if ch != "0":
            return False
    return True


def _cmp_operands(self, other, equality_op):
    if isinstance(other, Decimal):
        return (self, other)
    if isinstance(other, int):
        return (self, Decimal(other))
    if isinstance(other, _Fraction):
        a = self
        if not a._is_special:
            a = _dec_from_triple(a._sign, str(int(a._int) * other.denominator), a._exp)
        return (a, Decimal(other.numerator))
    if equality_op and isinstance(other, complex) and other.imag == 0:
        other = other.real
    if isinstance(other, float):
        context = getcontext()
        if equality_op:
            context.flags[FloatOperation] = True
        else:
            context._raise_error(FloatOperation, "strict semantics for mixing floats and Decimals are enabled")
        return (self, Decimal.from_float(other))
    return None


def _convert_other(other, raiseit=False, allow_float=False):
    if isinstance(other, Decimal):
        return other
    if isinstance(other, int):
        return Decimal(other)
    if allow_float and isinstance(other, float):
        return Decimal.from_float(other)
    if raiseit:
        raise TypeError("conversion from " + type(other).__name__ + " to Decimal is not supported")
    return NotImplemented


class _WorkRep:
    def __init__(self, value=None):
        if value is None:
            self.sign = 0
            self.int = 0
            self.exp = 0
        else:
            self.sign = value._sign
            self.int = int(value._int)
            self.exp = value._exp


def _normalize(op1, op2, prec=0):
    if op1.exp < op2.exp:
        tmp = op2
        other = op1
    else:
        tmp = op1
        other = op2
    tmp_len = len(str(tmp.int))
    other_len = len(str(other.int))
    exp = tmp.exp + min(-1, tmp_len - prec - 2)
    if other_len + other.exp - 1 < exp:
        other.int = 1
        other.exp = exp
    tmp.int *= 10 ** (tmp.exp - other.exp)
    tmp.exp = other.exp
    return op1, op2


def _decimal_lshift_exact(n, e):
    if n == 0:
        return 0
    elif e >= 0:
        return n * 10 ** e
    else:
        str_n = str(abs(n))
        val_n = len(str_n) - len(str_n.rstrip("0"))
        if val_n < -e:
            return None
        return n // 10 ** -e


def _log10_lb(c):
    if c <= 0:
        raise ValueError("The argument to _log10_lb should be nonnegative.")
    str_c = str(c)
    correction = {"1": 100, "2": 70, "3": 53, "4": 40, "5": 31, "6": 23, "7": 16, "8": 10, "9": 5}
    return 100 * len(str_c) - correction[str_c[0]]


def _div_nearest(a, b):
    q, r = divmod(a, b)
    if 2 * r + (q & 1) > b:
        q += 1
    return q


def _sqrt_nearest(n, a):
    if n <= 0 or a <= 0:
        raise ValueError("Both arguments to _sqrt_nearest should be positive.")
    b = 0
    while a != b:
        b, a = a, a - -n // a >> 1
    return a


def _rshift_nearest(x, shift):
    b = 1 << shift
    q, r = divmod(x, b)
    if 2 * r + (q & 1) > b:
        q += 1
    return q


def _iexp(x, M, L=8):
    R = _nbits((x << L) // M)
    T = -int(-10 * len(str(M)) // (3 * L))
    y = _div_nearest(x, T)
    Mshift = M << R
    for i in range(T - 1, 0, -1):
        y = _div_nearest(x * (Mshift + y), Mshift * i)
    for k in range(R - 1, -1, -1):
        Mshift = M << (k + 2)
        y = _div_nearest(y * (y + Mshift), Mshift)
    return M + y


def _ilog(x, M, L=8):
    y = x - M
    R = 0
    while (R <= L and abs(y) << L - R >= M or R > L and abs(y) >> R - L >= M):
        y = _div_nearest((M * y) << 1, M + _sqrt_nearest(M * (M + _rshift_nearest(y, R)), M))
        R += 1
    T = -int(-10 * len(str(M)) // (3 * L))
    yshift = _rshift_nearest(y, R)
    w = _div_nearest(M, T)
    for k in range(T - 1, 0, -1):
        w = _div_nearest(M, k) - _div_nearest(yshift * w, M)
    return _div_nearest(w * y, M)


class _Log10State:
    def __init__(self):
        self.digits = "23025850929940456840179914546843642076011014886287729760333279009675726096773524802359972050895982983419677840422862486334095254650828067566662873690987816894829072083255546808437998948262331985283935053089653777326288461633662222876982198867465436674744042432743651550489343149393914699"


_log10_state = _Log10State()


def _log10_digits(p):
    if p < 0:
        raise ValueError("p should be nonnegative")
    st = _log10_state
    if p >= len(st.digits):
        extra = 3
        while True:
            M = 10 ** (p + extra + 2)
            digits = str(_div_nearest(_ilog(10 * M, M), 100))
            if digits[-extra:] != "0" * extra:
                break
            extra += 3
        st.digits = digits.rstrip("0")[:-1]
    return int(st.digits[:p + 1])


def _dexp(c, e, p):
    p += 2
    extra = max(0, e + len(str(c)) - 1)
    q = p + extra
    shift = e + q
    if shift >= 0:
        cshift = c * 10 ** shift
    else:
        cshift = c // 10 ** -shift
    quot, rem = divmod(cshift, _log10_digits(q))
    rem = _div_nearest(rem, 10 ** extra)
    return _div_nearest(_iexp(rem, 10 ** p), 1000), quot - p + 3


def _dlog10(c, e, p):
    p += 2
    l = len(str(c))
    f = e + l - (e + l >= 1)
    if p > 0:
        M = 10 ** p
        k = e + p - f
        if k >= 0:
            c *= 10 ** k
        else:
            c = _div_nearest(c, 10 ** -k)
        log_d = _ilog(c, M)
        log_10 = _log10_digits(p)
        log_d = _div_nearest(log_d * M, log_10)
        log_tenpower = f * M
    else:
        log_d = 0
        log_tenpower = _div_nearest(f, 10 ** -p)
    return _div_nearest(log_tenpower + log_d, 100)


def _dlog(c, e, p):
    p += 2
    l = len(str(c))
    f = e + l - (e + l >= 1)
    if p > 0:
        k = e + p - f
        if k >= 0:
            c *= 10 ** k
        else:
            c = _div_nearest(c, 10 ** -k)
        log_d = _ilog(c, 10 ** p)
    else:
        log_d = 0
    if f:
        extra = len(str(abs(f))) - 1
        if p + extra >= 0:
            f_log_ten = _div_nearest(f * _log10_digits(p + extra), 10 ** extra)
        else:
            f_log_ten = 0
    else:
        f_log_ten = 0
    return _div_nearest(f_log_ten + log_d, 100)


def _dpower(xc, xe, yc, ye, p):
    b = len(str(abs(yc))) + ye
    lxc = _dlog(xc, xe, p + b + 1)
    shift = ye - b
    if shift >= 0:
        pc = lxc * yc * 10 ** shift
    else:
        pc = _div_nearest(lxc * yc, 10 ** -shift)
    if pc == 0:
        if (len(str(xc)) + xe >= 1) == (yc > 0):
            coeff, exp = 10 ** (p - 1) + 1, 1 - p
        else:
            coeff, exp = 10 ** p - 1, -p
    else:
        coeff, exp = _dexp(pc, -(p + 1), p + 1)
        coeff = _div_nearest(coeff, 10)
        exp += 1
    return coeff, exp


_Infinity = Decimal("Inf")
_NegativeInfinity = Decimal("-Inf")
_NaN = Decimal("NaN")
_Zero = Decimal(0)
_One = Decimal(1)
_NegativeOne = Decimal(-1)
_SignedInfinity = [_Infinity, _NegativeInfinity]


def _handle_error(condition, context, args):
    if condition is InvalidOperation or condition is ConversionSyntax or condition is DivisionImpossible or condition is DivisionUndefined or condition is InvalidContext:
        if len(args) > 0:
            ans = _dec_from_triple(args[0]._sign, args[0]._int, "n", True)
            return ans._fix_nan(context)
        return _NaN
    if condition is DivisionByZero:
        return _SignedInfinity[args[0]]
    if condition is Overflow:
        sign = args[0]
        if context.rounding in (ROUND_HALF_UP, ROUND_HALF_EVEN, ROUND_HALF_DOWN, ROUND_UP):
            return _SignedInfinity[sign]
        if sign == 0:
            if context.rounding == ROUND_CEILING:
                return _SignedInfinity[sign]
            return _dec_from_triple(sign, "9" * context.prec, context.Emax - context.prec + 1)
        if context.rounding == ROUND_FLOOR:
            return _SignedInfinity[sign]
        return _dec_from_triple(sign, "9" * context.prec, context.Emax - context.prec + 1)
    return None


def _condition_root(condition):
    if condition is ConversionSyntax or condition is DivisionImpossible or condition is DivisionUndefined or condition is InvalidContext:
        return InvalidOperation
    return condition


def _signal_message(error):
    names = [error]
    return "[" + ", ".join(["<class 'decimal." + n.__name__ + "'>" for n in names]) + "]"


class Context:
    def __init__(self, prec=None, rounding=None, Emin=None, Emax=None, capitals=None, clamp=None, flags=None, traps=None, _ignored_flags=None):
        dc = _cstate.default
        if dc is None:
            self.prec = 28
            self.rounding = ROUND_HALF_EVEN
            self.Emin = -999999
            self.Emax = 999999
            self.capitals = 1
            self.clamp = 0
            self.traps = {}
            for s in _signals:
                self.traps[s] = (s is InvalidOperation or s is DivisionByZero or s is Overflow)
        else:
            self.prec = dc.prec
            self.rounding = dc.rounding
            self.Emin = dc.Emin
            self.Emax = dc.Emax
            self.capitals = dc.capitals
            self.clamp = dc.clamp
            self.traps = dict(dc.traps)
        if prec is not None:
            self.prec = prec
        if rounding is not None:
            self.rounding = rounding
        if Emin is not None:
            self.Emin = Emin
        if Emax is not None:
            self.Emax = Emax
        if capitals is not None:
            self.capitals = capitals
        if clamp is not None:
            self.clamp = clamp
        if traps is not None:
            if isinstance(traps, dict):
                self.traps = dict(traps)
                for s in _signals:
                    if s not in self.traps:
                        self.traps[s] = False
            else:
                self.traps = {}
                for s in _signals:
                    self.traps[s] = s in traps
        self.flags = {}
        for s in _signals:
            self.flags[s] = False
        if flags is not None:
            if isinstance(flags, dict):
                for s in flags:
                    self.flags[s] = flags[s]
            else:
                for s in flags:
                    self.flags[s] = True
        self._ignored_flags = []
        if self.prec < 1 or self.prec > MAX_PREC:
            raise ValueError("valid range for prec is [1, MAX_PREC]")
        if self.Emin > 0 or self.Emin < MIN_EMIN:
            raise ValueError("valid range for Emin is [MIN_EMIN, 0]")
        if self.Emax < 0 or self.Emax > MAX_EMAX:
            raise ValueError("valid range for Emax is [0, MAX_EMAX]")

    def __repr__(self):
        names = [f.__name__ for f in _repr_order if self.flags[f]]
        tnames = [t.__name__ for t in _repr_order if self.traps[t]]
        return "Context(prec=" + str(self.prec) + ", rounding=" + self.rounding + ", Emin=" + str(self.Emin) + ", Emax=" + str(self.Emax) + ", capitals=" + str(self.capitals) + ", clamp=" + str(self.clamp) + ", flags=[" + ", ".join(names) + "], traps=[" + ", ".join(tnames) + "])"

    def clear_flags(self):
        for flag in self.flags:
            self.flags[flag] = False

    def clear_traps(self):
        for flag in self.traps:
            self.traps[flag] = False

    def _shallow_copy(self):
        nc = Context(self.prec, self.rounding, self.Emin, self.Emax, self.capitals, self.clamp, self.flags, self.traps)
        nc._ignored_flags = self._ignored_flags
        return nc

    def copy(self):
        nc = Context(self.prec, self.rounding, self.Emin, self.Emax, self.capitals, self.clamp, self.flags, self.traps)
        nc._ignored_flags = list(self._ignored_flags)
        return nc

    def _raise_error(self, condition, explanation=None, *args):
        error = _condition_root(condition)
        if error in self._ignored_flags:
            return _handle_error(condition, self, args)
        self.flags[error] = True
        if not self.traps[error]:
            return _handle_error(condition, self, args)
        raise error(_signal_message(condition))

    def _ignore_all_flags(self):
        self._ignored_flags = list(_signals)

    def _ignore_flags(self, *flags):
        self._ignored_flags = self._ignored_flags + list(flags)
        return list(flags)

    def _regard_flags(self, *flags):
        for flag in flags:
            self._ignored_flags.remove(flag)

    def Etiny(self):
        return int(self.Emin - self.prec + 1)

    def Etop(self):
        return int(self.Emax - self.prec + 1)

    def _set_rounding(self, type):
        rounding = self.rounding
        self.rounding = type
        return rounding

    def create_decimal(self, num="0"):
        if isinstance(num, str) and (num != num.strip() or "_" in num):
            return self._raise_error(ConversionSyntax, "trailing or leading whitespace and underscores are not permitted.")
        d = Decimal(num, context=self)
        if d._isnan() and len(d._int) > self.prec - self.clamp:
            return self._raise_error(ConversionSyntax, "diagnostic info too long in NaN")
        return d._fix(self)

    def create_decimal_from_float(self, f):
        d = Decimal.from_float(f)
        context = self.copy()
        context.clear_flags()
        return d._fix(self)

    def abs(self, a):
        a = _convert_other(a, raiseit=True)
        return a._abs(True, self)

    def add(self, a, b):
        a = _convert_other(a, raiseit=True)
        r = a._add(b, context=self)
        if r is NotImplemented:
            raise TypeError("conversion from " + type(b).__name__ + " to Decimal is not supported")
        return r

    def _apply(self, a, name):
        a = _convert_other(a, raiseit=True)
        return name

    def canonical(self, a):
        return a.canonical()

    def compare(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.compare(b, context=self)

    def compare_signal(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.compare_signal(b, context=self)

    def compare_total(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.compare_total(b)

    def compare_total_mag(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.compare_total_mag(b)

    def copy_abs(self, a):
        a = _convert_other(a, raiseit=True)
        return a.copy_abs()

    def copy_decimal(self, a):
        a = _convert_other(a, raiseit=True)
        return Decimal(a)

    def copy_negate(self, a):
        a = _convert_other(a, raiseit=True)
        return a.copy_negate()

    def copy_sign(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.copy_sign(b)

    def divide(self, a, b):
        a = _convert_other(a, raiseit=True)
        r = a._truediv(b, context=self)
        if r is NotImplemented:
            raise TypeError("conversion from " + type(b).__name__ + " to Decimal is not supported")
        return r

    def divide_int(self, a, b):
        a = _convert_other(a, raiseit=True)
        r = a._floordiv(b, context=self)
        if r is NotImplemented:
            raise TypeError("conversion from " + type(b).__name__ + " to Decimal is not supported")
        return r

    def divmod(self, a, b):
        a = _convert_other(a, raiseit=True)
        r = a._divmod(b, context=self)
        if r is NotImplemented:
            raise TypeError("conversion from " + type(b).__name__ + " to Decimal is not supported")
        return r

    def exp(self, a):
        a = _convert_other(a, raiseit=True)
        return a.exp(context=self)

    def fma(self, a, b, c):
        a = _convert_other(a, raiseit=True)
        return a.fma(b, c, context=self)

    def is_canonical(self, a):
        return a.is_canonical()

    def is_finite(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_finite()

    def is_infinite(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_infinite()

    def is_nan(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_nan()

    def is_normal(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_normal(context=self)

    def is_qnan(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_qnan()

    def is_signed(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_signed()

    def is_snan(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_snan()

    def is_subnormal(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_subnormal(context=self)

    def is_zero(self, a):
        a = _convert_other(a, raiseit=True)
        return a.is_zero()

    def ln(self, a):
        a = _convert_other(a, raiseit=True)
        return a.ln(context=self)

    def log10(self, a):
        a = _convert_other(a, raiseit=True)
        return a.log10(context=self)

    def logb(self, a):
        a = _convert_other(a, raiseit=True)
        return a.logb(context=self)

    def logical_and(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.logical_and(b, context=self)

    def logical_invert(self, a):
        a = _convert_other(a, raiseit=True)
        return a.logical_invert(context=self)

    def logical_or(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.logical_or(b, context=self)

    def logical_xor(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.logical_xor(b, context=self)

    def max(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.max(b, context=self)

    def max_mag(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.max_mag(b, context=self)

    def min(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.min(b, context=self)

    def min_mag(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.min_mag(b, context=self)

    def minus(self, a):
        a = _convert_other(a, raiseit=True)
        return a.copy_negate_ctx(self)

    def multiply(self, a, b):
        a = _convert_other(a, raiseit=True)
        r = a._mul(b, context=self)
        if r is NotImplemented:
            raise TypeError("conversion from " + type(b).__name__ + " to Decimal is not supported")
        return r

    def next_minus(self, a):
        a = _convert_other(a, raiseit=True)
        return a.next_minus(context=self)

    def next_plus(self, a):
        a = _convert_other(a, raiseit=True)
        return a.next_plus(context=self)

    def next_toward(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.next_toward(b, context=self)

    def normalize(self, a):
        a = _convert_other(a, raiseit=True)
        return a.normalize(context=self)

    def number_class(self, a):
        a = _convert_other(a, raiseit=True)
        return a.number_class(context=self)

    def plus(self, a):
        a = _convert_other(a, raiseit=True)
        return a._pos(self)

    def power(self, a, b, modulo=None):
        a = _convert_other(a, raiseit=True)
        r = a._pow(b, modulo, context=self)
        if r is NotImplemented:
            raise TypeError("conversion from " + type(b).__name__ + " to Decimal is not supported")
        return r

    def quantize(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.quantize(b, context=self)

    def radix(self):
        return Decimal(10)

    def remainder(self, a, b):
        a = _convert_other(a, raiseit=True)
        r = a._mod(b, context=self)
        if r is NotImplemented:
            raise TypeError("conversion from " + type(b).__name__ + " to Decimal is not supported")
        return r

    def remainder_near(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.remainder_near(b, context=self)

    def rotate(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.rotate(b, context=self)

    def same_quantum(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.same_quantum(b)

    def scaleb(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.scaleb(b, context=self)

    def shift(self, a, b):
        a = _convert_other(a, raiseit=True)
        return a.shift(b, context=self)

    def sqrt(self, a):
        a = _convert_other(a, raiseit=True)
        return a.sqrt(context=self)

    def subtract(self, a, b):
        a = _convert_other(a, raiseit=True)
        r = a._sub(b, context=self)
        if r is NotImplemented:
            raise TypeError("conversion from " + type(b).__name__ + " to Decimal is not supported")
        return r

    def to_eng_string(self, a):
        a = _convert_other(a, raiseit=True)
        return a.to_eng_string(context=self)

    def to_sci_string(self, a):
        a = _convert_other(a, raiseit=True)
        return a._str(False, self)

    def to_integral_exact(self, a):
        a = _convert_other(a, raiseit=True)
        return a.to_integral_exact(context=self)

    def to_integral_value(self, a):
        a = _convert_other(a, raiseit=True)
        return a.to_integral_value(context=self)

    def to_integral(self, a):
        a = _convert_other(a, raiseit=True)
        return a.to_integral_value(context=self)


DefaultContext = Context(prec=28, rounding=ROUND_HALF_EVEN, traps=[DivisionByZero, Overflow, InvalidOperation], flags=[], Emax=999999, Emin=-999999, capitals=1, clamp=0)
BasicContext = Context(prec=9, rounding=ROUND_HALF_UP, traps=[DivisionByZero, Overflow, InvalidOperation, Clamped, Underflow], flags=[])
ExtendedContext = Context(prec=9, rounding=ROUND_HALF_EVEN, traps=[], flags=[])


import numbers as _numbers
_numbers.Number.register(Decimal)
