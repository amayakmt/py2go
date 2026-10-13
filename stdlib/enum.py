"""enum: enumerations.

A class that derives from Enum is rewritten when the program is compiled:
its members become module variables (Color.RED is the variable of the member)
and Color(value), Color[name] and iteration over Color are calls of functions
the compiler writes. The classes here hold what the members share.
"""


class auto:
    pass


class Enum:
    def __init__(self, *args):
        pass

    @property
    def name(self):
        return self._name_

    @property
    def value(self):
        return self._value_

    def __repr__(self):
        return "<" + type(self).__name__ + "." + self._name_ + ": " + repr(self._value_) + ">"

    def __str__(self):
        return type(self).__name__ + "." + self._name_

    def __format__(self, spec):
        return format(str(self), spec)

    def __hash__(self):
        return hash(self._name_)


class IntEnum(Enum):
    def __int__(self):
        return self._value_

    def __index__(self):
        return self._value_

    def __str__(self):
        return str(self._value_)

    def __format__(self, spec):
        return format(self._value_, spec)

    def __hash__(self):
        return hash(self._value_)

    def __eq__(self, other):
        if isinstance(other, Enum):
            return self._value_ == other._value_
        return self._value_ == other

    def __ne__(self, other):
        return not (self == other)

    def __lt__(self, other):
        return self._value_ < _val(other)

    def __le__(self, other):
        return self._value_ <= _val(other)

    def __gt__(self, other):
        return self._value_ > _val(other)

    def __ge__(self, other):
        return self._value_ >= _val(other)

    def __add__(self, other):
        return self._value_ + _val(other)

    def __radd__(self, other):
        return _val(other) + self._value_

    def __sub__(self, other):
        return self._value_ - _val(other)

    def __rsub__(self, other):
        return _val(other) - self._value_

    def __mul__(self, other):
        return self._value_ * _val(other)

    def __rmul__(self, other):
        return _val(other) * self._value_

    def __floordiv__(self, other):
        return self._value_ // _val(other)

    def __truediv__(self, other):
        return self._value_ / _val(other)

    def __mod__(self, other):
        return self._value_ % _val(other)

    def __neg__(self):
        return -self._value_

    def __abs__(self):
        return abs(self._value_)

    def __bool__(self):
        return self._value_ != 0

    def __and__(self, other):
        return self._value_ & _val(other)

    def __or__(self, other):
        return self._value_ | _val(other)

    def __xor__(self, other):
        return self._value_ ^ _val(other)

    def __lshift__(self, other):
        return self._value_ << _val(other)

    def __rshift__(self, other):
        return self._value_ >> _val(other)


class StrEnum(Enum):
    def __str__(self):
        return self._value_

    def __format__(self, spec):
        return format(self._value_, spec)

    def __hash__(self):
        return hash(self._value_)

    def __eq__(self, other):
        if isinstance(other, Enum):
            return self._value_ == other._value_
        return self._value_ == other

    def __ne__(self, other):
        return not (self == other)

    def __lt__(self, other):
        return self._value_ < _val(other)

    def __len__(self):
        return len(self._value_)

    def __contains__(self, item):
        return item in self._value_

    def upper(self):
        return self._value_.upper()

    def lower(self):
        return self._value_.lower()

    def __add__(self, other):
        return self._value_ + _val(other)


def _val(x):
    if isinstance(x, Enum):
        return x._value_
    return x


def _flag_names(cls_members, value):
    names = []
    for m in cls_members:
        if m._value_ and m._value_ & value == m._value_ and (m._value_ & (m._value_ - 1)) == 0:
            names.append(m._name_)
    return "|".join(names)


class Flag(Enum):
    def __repr__(self):
        if not self._name_:
            return "<" + type(self).__name__ + ": " + repr(self._value_) + ">"
        return "<" + type(self).__name__ + "." + self._name_ + ": " + repr(self._value_) + ">"

    def __str__(self):
        if not self._name_:
            return type(self).__name__ + "(" + str(self._value_) + ")"
        return type(self).__name__ + "." + self._name_

    def __hash__(self):
        return hash(self._value_)

    def __eq__(self, other):
        return isinstance(other, Flag) and type(self).__name__ == type(other).__name__ and self._value_ == other._value_

    def __ne__(self, other):
        return not (self == other)

    def __bool__(self):
        return self._value_ != 0

    def __or__(self, other):
        return self._lookup(self._value_ | _val(other))

    def __and__(self, other):
        return self._lookup(self._value_ & _val(other))

    def __xor__(self, other):
        return self._lookup(self._value_ ^ _val(other))

    def __invert__(self):
        return self._lookup(self._all_bits() & ~self._value_)

    def __contains__(self, other):
        return other._value_ & self._value_ == other._value_

    def __iter__(self):
        return iter([m for m in self._members() if m._value_ and m._value_ & self._value_ == m._value_ and (m._value_ & (m._value_ - 1)) == 0])

    def __len__(self):
        return len(list(self.__iter__()))


class IntFlag(Flag):
    def __int__(self):
        return self._value_

    def __index__(self):
        return self._value_

    def __str__(self):
        return str(self._value_)

    def __format__(self, spec):
        return format(self._value_, spec)

    def __eq__(self, other):
        return self._value_ == _val(other)

    def __lt__(self, other):
        return self._value_ < _val(other)

    def __le__(self, other):
        return self._value_ <= _val(other)

    def __gt__(self, other):
        return self._value_ > _val(other)

    def __ge__(self, other):
        return self._value_ >= _val(other)

    def __or__(self, other):
        return self._lookup(self._value_ | _val(other))

    def __ror__(self, other):
        return self._lookup(self._value_ | _val(other))

    def __and__(self, other):
        return self._lookup(self._value_ & _val(other))

    def __rand__(self, other):
        return self._lookup(self._value_ & _val(other))


def unique(enumeration):
    return enumeration
