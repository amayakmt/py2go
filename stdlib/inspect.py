"""Get information about live objects (reduced).

Predicates, getmembers, cleandoc, and signatures built from what the program
knows of a function value: parameter names, kinds, defaults. There are no
annotations, source text, frames or code objects."""

__all__ = ['isclass', 'isfunction', 'ismethod', 'isbuiltin', 'ismodule', 'isroutine', 'iscoroutinefunction', 'isgeneratorfunction', 'isgenerator', 'iscoroutine', 'isawaitable', 'isabstract', 'getmembers', 'getdoc', 'cleandoc', 'signature', 'Signature', 'Parameter', 'BoundArguments', 'getmro', 'ismethoddescriptor', 'isdatadescriptor', 'currentframe', 'stack']


def _obj(x: object) -> object:
    return x


def isclass(object):
    """Return true if the object is a class."""
    return isinstance(object, type)


def ismodule(object):
    return False


def isfunction(object):
    """Return true if the object is a user-defined function (or method)."""
    if isinstance(object, type):
        return False
    return _native("inspect.sigdata", object) is not None


def ismethod(object):
    return False


def isbuiltin(object):
    return callable(object) and not isinstance(object, type) and _native("inspect.sigdata", object) is None


def isroutine(object):
    return isbuiltin(object) or isfunction(object) or ismethod(object)


def ismethoddescriptor(object):
    return False


def isdatadescriptor(object):
    return False


def isabstract(object):
    return False


def iscoroutinefunction(obj):
    data = _native("inspect.sigdata", obj)
    return data is not None and data[4]


def isgeneratorfunction(obj):
    return False


def isgenerator(object):
    return hasattr(object, '__next__') and hasattr(object, 'send') and hasattr(object, 'throw')


def iscoroutine(object):
    return False


def isawaitable(object):
    return hasattr(object, '__await__')


def getmro(cls):
    """Return tuple of base classes (including cls) in method resolution order."""
    return tuple(cls.__mro__) if hasattr(cls, '__mro__') else (cls,)


def currentframe():
    return None


def stack(context=1):
    return []


def getdoc(object):
    return None


def cleandoc(doc):
    """Clean up indentation from docstrings."""
    lines = doc.expandtabs().split('\n')
    margin = 1 << 30
    for line in lines[1:]:
        content = len(line.lstrip())
        if content:
            indent = len(line) - content
            margin = min(margin, indent)
    if lines:
        lines[0] = lines[0].lstrip()
    if margin < (1 << 30):
        for i in range(1, len(lines)):
            lines[i] = lines[i][margin:]
    while lines and (not lines[-1]):
        lines.pop()
    while lines and (not lines[0]):
        lines.pop(0)
    return '\n'.join(lines)


def getmembers(object, predicate=None):
    """Return all members of an object as (name, value) pairs sorted by name.
    Optionally, only return members that satisfy a given predicate."""
    results = []
    for key in dir(object):
        try:
            value = getattr(object, key)
        except AttributeError:
            continue
        if not predicate or predicate(value):
            results.append((key, value))
    results.sort(key=lambda pair: pair[0])
    return results


class _empty:
    """Marker object for Signature.empty and Parameter.empty."""


class Parameter:
    POSITIONAL_ONLY = 0
    POSITIONAL_OR_KEYWORD = 1
    VAR_POSITIONAL = 2
    KEYWORD_ONLY = 3
    VAR_KEYWORD = 4
    empty = _empty

    def __init__(self, name, kind, *, default=_empty, annotation=_empty):
        self._name = name
        self._kind = kind
        self._default = _obj(default)
        self._annotation = _obj(annotation)

    @property
    def name(self):
        return self._name

    @property
    def kind(self):
        return self._kind

    @property
    def default(self):
        return self._default

    @property
    def annotation(self):
        return self._annotation

    def __str__(self):
        kind = self._kind
        formatted = self._name
        if self._default is not _empty:
            formatted = '{}={}'.format(formatted, repr(self._default))
        if kind == 2:
            formatted = '*' + formatted
        elif kind == 4:
            formatted = '**' + formatted
        return formatted

    def __repr__(self):
        return '<Parameter "{}">'.format(self)

    def __eq__(self, other):
        if not isinstance(other, Parameter):
            return NotImplemented
        return self._name == other._name and self._kind == other._kind and self._default == other._default


class BoundArguments:

    def __init__(self, signature, arguments):
        self.signature = signature
        self.arguments = arguments

    @property
    def args(self):
        out = []
        for name in self.arguments:
            p = self.signature.parameters[name]
            if p.kind == Parameter.VAR_POSITIONAL:
                out.extend(self.arguments[name])
            elif p.kind == Parameter.POSITIONAL_OR_KEYWORD:
                out.append(self.arguments[name])
        return tuple(out)

    @property
    def kwargs(self):
        out = {}
        for name in self.arguments:
            p = self.signature.parameters[name]
            if p.kind == Parameter.KEYWORD_ONLY:
                out[name] = self.arguments[name]
            elif p.kind == Parameter.VAR_KEYWORD:
                out.update(self.arguments[name])
        return out

    def apply_defaults(self):
        for name in self.signature.parameters:
            if name not in self.arguments:
                p = self.signature.parameters[name]
                if p.default is not _empty:
                    self.arguments[name] = p.default
                elif p.kind == Parameter.VAR_POSITIONAL:
                    self.arguments[name] = ()
                elif p.kind == Parameter.VAR_KEYWORD:
                    self.arguments[name] = {}
        ordered = {}
        for name in self.signature.parameters:
            if name in self.arguments:
                ordered[name] = self.arguments[name]
        self.arguments = ordered


class Signature:
    empty = _empty

    def __init__(self, parameters=None, *, return_annotation=_empty):
        self._params = _obj({})
        if parameters is not None:
            for p in parameters:
                self._params[p.name] = p
        self._return_annotation = _obj(return_annotation)

    @property
    def parameters(self):
        return self._params

    @property
    def return_annotation(self):
        return self._return_annotation

    def __str__(self):
        result = []
        render_kw_only_separator = True
        for param in self._params.values():
            formatted = str(param)
            kind = param.kind
            if kind == Parameter.VAR_POSITIONAL:
                render_kw_only_separator = False
            elif kind == Parameter.KEYWORD_ONLY and render_kw_only_separator:
                result.append('*')
                render_kw_only_separator = False
            result.append(formatted)
        return '(' + ', '.join(result) + ')'

    def __repr__(self):
        return '<Signature {}>'.format(self)

    def bind(self, *args, **kwargs):
        arguments = {}
        names = [n for n in self._params]
        pos_names = [n for n in names if self._params[n].kind == Parameter.POSITIONAL_OR_KEYWORD]
        var_pos = None
        var_kw = None
        for n in names:
            if self._params[n].kind == Parameter.VAR_POSITIONAL:
                var_pos = n
            elif self._params[n].kind == Parameter.VAR_KEYWORD:
                var_kw = n
        for i, value in enumerate(args):
            if i < len(pos_names):
                arguments[pos_names[i]] = value
            elif var_pos is not None:
                arguments[var_pos] = tuple(args[len(pos_names):])
                break
            else:
                raise TypeError('too many positional arguments')
        extra = {}
        for key in kwargs:
            if key in self._params and self._params[key].kind != Parameter.VAR_POSITIONAL and self._params[key].kind != Parameter.VAR_KEYWORD:
                if key in arguments:
                    raise TypeError('multiple values for argument %r' % key)
                arguments[key] = kwargs[key]
            elif var_kw is not None:
                extra[key] = kwargs[key]
            else:
                raise TypeError('got an unexpected keyword argument %r' % key)
        if var_kw is not None and extra:
            arguments[var_kw] = extra
        for n in names:
            p = self._params[n]
            if n not in arguments and p.default is _empty and p.kind != Parameter.VAR_POSITIONAL and p.kind != Parameter.VAR_KEYWORD:
                raise TypeError('missing a required argument: %r' % n)
        ordered = {}
        for n in names:
            if n in arguments:
                ordered[n] = arguments[n]
        return BoundArguments(self, ordered)


def signature(obj, *, follow_wrapped=True, globals=None, locals=None, eval_str=False):
    """Get a signature object for the passed callable."""
    data = _native("inspect.sigdata", obj)
    if data is None:
        raise ValueError('no signature found for builtin %r' % (obj,))
    params = []
    for entry in data[1]:
        name, kind, has_default, default = entry
        pk = Parameter.KEYWORD_ONLY if kind == 'kwonly' else Parameter.POSITIONAL_OR_KEYWORD
        if has_default:
            params.append(Parameter(name, pk, default=default))
        else:
            params.append(Parameter(name, pk))
    # *args goes after the positional parameters, **kwargs last.
    ordered = [p for p in params if p.kind == Parameter.POSITIONAL_OR_KEYWORD]
    if data[2]:
        ordered.append(Parameter(data[2], Parameter.VAR_POSITIONAL))
    ordered.extend([p for p in params if p.kind == Parameter.KEYWORD_ONLY])
    if data[3]:
        ordered.append(Parameter(data[3], Parameter.VAR_KEYWORD))
    return Signature(ordered)
