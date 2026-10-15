"""Mock objects (reduced hand-written version of unittest.mock).

Mock, MagicMock (with the usual defaults for the common magic methods),
call, ANY, sentinel, DEFAULT, patch.object and patch.dict as context
managers. No patching by dotted name (modules are not objects here), no
autospec, no mock_open."""
__all__ = ['Mock', 'MagicMock', 'NonCallableMock', 'call', 'ANY', 'DEFAULT', 'sentinel', 'patch', 'create_autospec']


def _obj(x: object) -> object:
    return x


class _SentinelObject(object):

    def __init__(self, name):
        self.name = name

    def __repr__(self):
        return 'sentinel.%s' % self.name


class _Sentinels(object):

    def __init__(self):
        self._created = _obj({})

    def __getattr__(self, name):
        if name == '__bases__':
            raise AttributeError
        if name not in self._created:
            self._created[name] = _SentinelObject(name)
        return self._created[name]


sentinel = _Sentinels()
DEFAULT = sentinel.DEFAULT


class _ANY(object):
    "A helper object that compares equal to everything."

    def __eq__(self, other):
        return True

    def __ne__(self, other):
        return False

    def __repr__(self):
        return '<ANY>'


ANY = _ANY()


def _format_call_signature(name, args, kwargs):
    message = '%s(%%s)' % name
    formatted_args = ''
    args_string = ', '.join([repr(arg) for arg in args])
    kwargs_string = ', '.join(['%s=%r' % (key, kwargs[key]) for key in kwargs])
    if args_string:
        formatted_args = args_string
    if kwargs_string:
        if formatted_args:
            formatted_args += ', '
        formatted_args += kwargs_string
    return message % formatted_args


class _Call(object):
    """A recorded call: a name, positional arguments and keyword arguments.
    Compares equal to another call with the same arguments (ANY matches)."""

    def __init__(self, name='', args=(), kwargs=None):
        self._mock_name = name
        self.args = tuple(args)
        self.kwargs = _obj(kwargs if kwargs is not None else {})

    def __eq__(self, other):
        if not isinstance(other, _Call):
            return False
        if self._mock_name != '' and other._mock_name != self._mock_name:
            return False
        return self.args == other.args and self.kwargs == other.kwargs

    def __call__(self, *args, **kwargs):
        return _Call(self._mock_name, args, kwargs)

    def __getattr__(self, attr):
        if attr.startswith('__'):
            raise AttributeError(attr)
        if self._mock_name == '':
            return _Call(attr)
        return _Call(self._mock_name + '.' + attr)

    def __repr__(self):
        if self._mock_name == '':
            return _format_call_signature('call', self.args, self.kwargs)
        name = self._mock_name
        if name.startswith('()'):
            name = 'call' + name
        else:
            name = 'call.' + name
        return _format_call_signature(name, self.args, self.kwargs)


call = _Call('')


class NonCallableMock(object):
    """A non-callable version of `Mock`"""

    def __init__(self, spec=None, side_effect=None, return_value=DEFAULT, name=None, _parent=None, _parent_key='', **kwargs):
        self._mock_children = _obj({})
        self._mock_name = _obj(name)
        self._mock_parent = _obj(_parent)
        self._mock_parent_key = _parent_key
        self._mock_spec = _obj(None)
        if spec is not None:
            if isinstance(spec, list):
                self._mock_spec = _obj(list(spec))
            else:
                self._mock_spec = _obj(dir(spec))
        self.called = False
        self.call_count = 0
        self.call_args = _obj(None)
        self.call_args_list = _obj([])
        self.mock_calls = _obj([])
        self.method_calls = _obj([])
        if return_value is not DEFAULT:
            self._mock_children['return_value'] = return_value
        if side_effect is not None:
            self._mock_children['side_effect'] = side_effect
        for key in kwargs:
            self._mock_children[key] = kwargs[key]

    def __getattr__(self, name):
        if name.startswith('_mock_') or name == '__bases__' or (name.startswith('__') and name.endswith('__')):
            raise AttributeError(name)
        if name in self._mock_children:
            return self._mock_children[name]
        if name == 'side_effect':
            return None
        if name == 'return_value':
            rv = Mock(name=None, _parent=self, _parent_key='()')
            self._mock_children['return_value'] = rv
            return rv
        if self._mock_spec is not None and name not in self._mock_spec:
            raise AttributeError("Mock object has no attribute '%s'" % name)
        child = Mock(name=name, _parent=self, _parent_key=name)
        self._mock_children[name] = child
        return child

    def __repr__(self):
        name = self._mock_name
        if name is None:
            name = 'mock'
        else:
            name = 'mock.' + name
        return "<%s name='%s' id='%s'>" % (type(self).__name__, name, id(self))

    def reset_mock(self, visited=None, *, return_value=False, side_effect=False):
        self.called = False
        self.call_args = None
        self.call_count = 0
        self.mock_calls = []
        self.call_args_list = []
        self.method_calls = []
        if return_value and 'return_value' in self._mock_children:
            del self._mock_children['return_value']
        if side_effect and 'side_effect' in self._mock_children:
            del self._mock_children['side_effect']
        for key in list(self._mock_children.keys()):
            child = self._mock_children[key]
            if isinstance(child, NonCallableMock):
                child.reset_mock()

    def configure_mock(self, **kwargs):
        for arg in kwargs:
            obj = self
            parts = arg.split('.')
            final = parts[-1]
            for entry in parts[:-1]:
                obj = getattr(obj, entry)
            setattr(obj, final, kwargs[arg])

    def _mock_name_or_mock(self):
        n = self._mock_name
        return n if n else 'mock'

    def _format_mock_call_signature(self, args, kwargs):
        return _format_call_signature(self._mock_name_or_mock(), args, kwargs)

    def _calls_repr(self):
        if not self.mock_calls:
            return ''
        return '\nCalls: %s.' % (repr(self.mock_calls),)

    def assert_not_called(self):
        if self.call_count != 0:
            msg = "Expected '%s' to not have been called. Called %s times.%s" % (self._mock_name_or_mock(), self.call_count, self._calls_repr())
            raise AssertionError(msg)

    def assert_called(self):
        if self.call_count == 0:
            msg = "Expected '%s' to have been called." % (self._mock_name_or_mock(),)
            raise AssertionError(msg)

    def assert_called_once(self):
        if not self.call_count == 1:
            msg = "Expected '%s' to have been called once. Called %s times.%s" % (self._mock_name_or_mock(), self.call_count, self._calls_repr())
            raise AssertionError(msg)

    def assert_called_with(self, *args, **kwargs):
        if self.call_args is None:
            expected = self._format_mock_call_signature(args, kwargs)
            raise AssertionError('expected call not found.\nExpected: %s\n  Actual: not called.' % expected)
        expected_call = _Call('', args, kwargs)
        actual_call = self.call_args
        if actual_call != expected_call:
            expected_string = self._format_mock_call_signature(args, kwargs)
            actual_string = self._format_mock_call_signature(actual_call.args, actual_call.kwargs)
            raise AssertionError('expected call not found.\nExpected: %s\n  Actual: %s' % (expected_string, actual_string))

    def assert_called_once_with(self, *args, **kwargs):
        if not self.call_count == 1:
            msg = "Expected '%s' to be called once. Called %s times.%s" % (self._mock_name_or_mock(), self.call_count, self._calls_repr())
            raise AssertionError(msg)
        return self.assert_called_with(*args, **kwargs)

    def assert_has_calls(self, calls, any_order=False):
        expected = [c for c in calls]
        all_calls = [c for c in self.mock_calls]
        if not any_order:
            if len(expected) == 0:
                return
            n = len(expected)
            for i in range(len(all_calls) - n + 1):
                if all_calls[i:i + n] == expected:
                    return
            raise AssertionError('Calls not found.\nExpected: %r%s' % (expected, self._calls_repr()))
        remaining = list(all_calls)
        not_found = []
        for kall in expected:
            found = False
            for i in range(len(remaining)):
                if remaining[i] == kall:
                    del remaining[i]
                    found = True
                    break
            if not found:
                not_found.append(kall)
        if not_found:
            raise AssertionError('%r does not contain all of %r in its call list, found %r instead' % (self._mock_name_or_mock(), expected, self.mock_calls))

    def assert_any_call(self, *args, **kwargs):
        expected_call = _Call('', args, kwargs)
        for c in self.call_args_list:
            if c == expected_call:
                return
        expected_string = self._format_mock_call_signature(args, kwargs)
        raise AssertionError('%s call not found' % expected_string)


class Mock(NonCallableMock):
    """Create a new `Mock` object. `Mock` takes several optional arguments
    that specify the behaviour of the Mock object."""

    def __call__(self, *args, **kwargs):
        self.called = True
        self.call_count += 1
        this_call = _Call('', args, kwargs)
        self.call_args = this_call
        self.call_args_list.append(this_call)
        self.mock_calls.append(_Call('', args, kwargs))
        node = self
        name = ''
        while node._mock_parent is not None:
            parent = node._mock_parent
            key = node._mock_parent_key
            if name == '':
                name = key
            elif name.startswith('()'):
                name = key + name
            else:
                name = key + '.' + name
            entry = _Call(name, args, kwargs)
            parent.mock_calls.append(entry)
            if not name.startswith('()'):
                parent.method_calls.append(entry)
            node = parent
        effect = self._mock_children.get('side_effect', None)
        if effect is not None:
            if isinstance(effect, BaseException):
                raise effect
            if isinstance(effect, type):
                raise effect()
            if callable(effect):
                result = effect(*args, **kwargs)
                if result is not DEFAULT:
                    return result
            else:
                it = iter(effect)
                self._mock_children['side_effect'] = it
                effect = it
        if effect is not None and hasattr(effect, '__next__'):
            result = next(effect)
            if isinstance(result, BaseException):
                raise result
            if result is not DEFAULT:
                return result
        return self.return_value


class MagicMock(Mock):
    """MagicMock is a subclass of Mock with default implementations of the
    common magic methods (__len__, __iter__, __bool__, __contains__,
    __getitem__, __int__, __str__, ...)."""

    def __len__(self):
        return 0

    def __bool__(self):
        return True

    def __iter__(self):
        return iter([])

    def __contains__(self, item):
        return False

    def __int__(self):
        return 1

    def __float__(self):
        return 1.0

    def __hash__(self):
        return id(self)

    def __getitem__(self, key):
        return self._item_mock()

    def __setitem__(self, key, value):
        pass

    def __delitem__(self, key):
        pass

    def __enter__(self):
        return self._item_mock()

    def __exit__(self, exc_type, exc_value, tb):
        return False

    def _item_mock(self):
        if '__item__' not in self._mock_children:
            self._mock_children['__item__'] = MagicMock(name=None, _parent=self, _parent_key='()')
        return self._mock_children['__item__']


def create_autospec(spec, spec_set=False, instance=False, **kwargs):
    return MagicMock(spec=spec, **kwargs)


class _patch_object(object):

    def __init__(self, target, attribute, new):
        self.target = _obj(target)
        self.attribute = attribute
        self.new = _obj(new)
        self.original = _obj(None)
        self.had = False

    def __enter__(self):
        self.had = hasattr(self.target, self.attribute)
        if self.had:
            self.original = getattr(self.target, self.attribute)
        new = self.new
        if new is DEFAULT:
            new = MagicMock(name=self.attribute)
        setattr(self.target, self.attribute, new)
        return new

    def __exit__(self, exc_type, exc_value, tb):
        if self.had:
            setattr(self.target, self.attribute, self.original)
        else:
            delattr(self.target, self.attribute)
        return False

    def start(self):
        return self.__enter__()

    def stop(self):
        return self.__exit__(None, None, None)


class _patch_dict(object):

    def __init__(self, in_dict, values=(), clear=False, **kwargs):
        self.in_dict = _obj(in_dict)
        self.values = _obj(dict(values))
        for k in kwargs:
            self.values[k] = kwargs[k]
        self.clear = clear
        self.original = _obj({})

    def __enter__(self):
        self.original = dict(self.in_dict)
        if self.clear:
            self.in_dict.clear()
        self.in_dict.update(self.values)
        return self.in_dict

    def __exit__(self, exc_type, exc_value, tb):
        self.in_dict.clear()
        self.in_dict.update(self.original)
        return False


class _patcher(object):

    def object(self, target, attribute, new=DEFAULT, **kwargs):
        if new is DEFAULT and len(kwargs) > 0:
            new = MagicMock(name=attribute, **kwargs)
        return _patch_object(target, attribute, new)

    def dict(self, in_dict, values=(), clear=False, **kwargs):
        return _patch_dict(in_dict, values, clear, **kwargs)

    def __call__(self, target, new=DEFAULT, **kwargs):
        raise NotImplementedError('patch() by dotted name is not supported; use patch.object')


patch = _patcher()
