from unittest import mock
from unittest.mock import ANY, MagicMock, Mock, call


def check(fn):
    try:
        fn()
    except AssertionError as e:
        return str(e)
    return "ok"


m = Mock()
print(m.called, m.call_count, m.call_args)
r = m(1, 2, key="v")
print(m.called, m.call_count, m.call_args, m.call_args_list)
print(r is m.return_value, m(3) is r)
m.assert_called()
print(check(lambda: m.assert_called_once()))
print(check(lambda: m.assert_called_with(9)))
m.assert_called_with(3)
m.assert_any_call(1, 2, key="v")
print(check(lambda: m.assert_any_call(5)))
print(check(lambda: m.assert_not_called()))
print(m.mock_calls)

m2 = Mock(return_value=42)
print(m2(), m2.return_value)
m3 = Mock(side_effect=[1, 2, ValueError("bad")])
print(m3(), m3())
try:
    m3()
except ValueError as e:
    print("raised", e)
m4 = Mock(side_effect=lambda x: x * 2)
print(m4(21))
m5 = Mock(side_effect=KeyError("k"))
try:
    m5()
except KeyError as e:
    print("raised", repr(e))

obj = Mock()
obj.method.return_value = "r"
print(obj.method(1), obj.method.call_args, obj.method_calls, obj.mock_calls)
obj.other.deep(2)
print(obj.mock_calls)
obj.method.assert_called_once_with(1)
print(check(lambda: obj.method.assert_called_once_with(2)))
obj().x(5)
print(obj.mock_calls[-2:])
obj.assert_has_calls([call.method(1), call.other.deep(2)])
print(check(lambda: obj.assert_has_calls([call.nothing()])))
print(call(1, a=2), call.foo(1), call.foo.bar(), call() == call())
print(Mock().anything == ANY, ANY == 3, call(ANY) == call(7))

obj.reset_mock()
print(obj.called, obj.mock_calls, obj.method.call_count)

mm = MagicMock()
print(len(mm), bool(mm), list(mm), 3 in mm, int(mm))
mm.attr.value = 10
print(mm.attr.value)
mm[0]
print(mm.configure_mock(a=1, **{"b.c": 2}), mm.a, mm.b.c)

spec = Mock(spec=["alpha"])
spec.alpha()
try:
    spec.beta
except AttributeError as e:
    print("AttributeError", e)


class Target:
    def __init__(self):
        self.value = 1


t = Target()
with mock.patch.object(t, "value", 99):
    print(t.value)
print(t.value)
d = {"a": 1}
with mock.patch.dict(d, {"b": 2}):
    print(sorted(d))
print(sorted(d))
print(mock.sentinel.thing, mock.DEFAULT)
