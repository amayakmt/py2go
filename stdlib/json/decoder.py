"""json.decoder"""
from builtins import json__decoder_pm_JSONDecodeError as JSONDecodeError


class JSONDecoder:
    def __init__(self, *, object_hook=None, parse_float=None, parse_int=None, parse_constant=None, strict=True, object_pairs_hook=None):
        self.object_hook = object_hook
        self.parse_float = parse_float
        self.parse_int = parse_int
        self.parse_constant = parse_constant
        self.object_pairs_hook = object_pairs_hook

    def decode(self, s, _w=None):
        return _native("json.loads", s, object_hook=self.object_hook, object_pairs_hook=self.object_pairs_hook, parse_float=self.parse_float, parse_int=self.parse_int, parse_constant=self.parse_constant)
