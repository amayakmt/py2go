"""json: JSON encoding and decoding."""
from .decoder import JSONDecoder, JSONDecodeError
from .encoder import JSONEncoder


def dumps(obj, *, skipkeys=False, ensure_ascii=True, check_circular=True, allow_nan=True, cls=None, indent=None, separators=None, default=None, sort_keys=False, **kw):
    if cls is None:
        cls = JSONEncoder
    return cls(skipkeys=skipkeys, ensure_ascii=ensure_ascii, check_circular=check_circular, allow_nan=allow_nan, indent=indent, separators=separators, default=default, sort_keys=sort_keys, **kw).encode(obj)


def dump(obj, fp, *, skipkeys=False, ensure_ascii=True, check_circular=True, allow_nan=True, cls=None, indent=None, separators=None, default=None, sort_keys=False, **kw):
    fp.write(dumps(obj, skipkeys=skipkeys, ensure_ascii=ensure_ascii, check_circular=check_circular, allow_nan=allow_nan, cls=cls, indent=indent, separators=separators, default=default, sort_keys=sort_keys, **kw))


def loads(s, *, cls=None, object_hook=None, parse_float=None, parse_int=None, parse_constant=None, object_pairs_hook=None, **kw):
    return _native("json.loads", s, object_hook=object_hook, object_pairs_hook=object_pairs_hook, parse_float=parse_float, parse_int=parse_int, parse_constant=parse_constant)


def load(fp, *, cls=None, object_hook=None, parse_float=None, parse_int=None, parse_constant=None, object_pairs_hook=None, **kw):
    return loads(fp.read(), cls=cls, object_hook=object_hook, parse_float=parse_float, parse_int=parse_int, parse_constant=parse_constant, object_pairs_hook=object_pairs_hook, **kw)
