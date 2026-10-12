from typing import Optional

eg: ExceptionGroup = ExceptionGroup("many", [ValueError("a"), TypeError("b"), ValueError("c")])
print(eg, repr(eg))
print(eg.message, len(eg.exceptions), [type(x).__name__ for x in eg.exceptions])
m, rest = eg.split(ValueError)
print(m, rest)
sub: Optional[ExceptionGroup] = eg.subgroup(TypeError)
print(sub, sub.exceptions[0] if sub is not None else None)
print(eg.subgroup(KeyError))
nested: ExceptionGroup = ExceptionGroup("outer", [eg, KeyError("k")])
only_values, others = nested.split(ValueError)
print(only_values, others)
print(isinstance(eg, Exception), isinstance(eg, ExceptionGroup), isinstance(eg, BaseExceptionGroup))
