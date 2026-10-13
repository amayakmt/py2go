"""pickle: Python object serialization (protocols 0-5)."""

import copyreg
import _compat_pickle
import struct

format_version = "4.0"
compatible_formats = ["1.0", "1.1", "1.2", "1.3", "2.0", "3.0", "4.0", "5.0"]
HIGHEST_PROTOCOL = 5
DEFAULT_PROTOCOL = 4

MARK = b"("
STOP = b"."
POP = b"0"
POP_MARK = b"1"
DUP = b"2"
FLOAT = b"F"
INT = b"I"
BININT = b"J"
BININT1 = b"K"
LONG = b"L"
BININT2 = b"M"
NONE = b"N"
PERSID = b"P"
BINPERSID = b"Q"
REDUCE = b"R"
STRING = b"S"
BINSTRING = b"T"
SHORT_BINSTRING = b"U"
UNICODE = b"V"
BINUNICODE = b"X"
APPEND = b"a"
BUILD = b"b"
GLOBAL = b"c"
DICT = b"d"
EMPTY_DICT = b"}"
APPENDS = b"e"
GET = b"g"
BINGET = b"h"
INST = b"i"
LONG_BINGET = b"j"
LIST = b"l"
EMPTY_LIST = b"]"
OBJ = b"o"
PUT = b"p"
BINPUT = b"q"
LONG_BINPUT = b"r"
SETITEM = b"s"
TUPLE = b"t"
EMPTY_TUPLE = b")"
SETITEMS = b"u"
BINFLOAT = b"G"
TRUE = b"I01\n"
FALSE = b"I00\n"
PROTO = b"\x80"
NEWOBJ = b"\x81"
EXT1 = b"\x82"
EXT2 = b"\x83"
EXT4 = b"\x84"
TUPLE1 = b"\x85"
TUPLE2 = b"\x86"
TUPLE3 = b"\x87"
NEWTRUE = b"\x88"
NEWFALSE = b"\x89"
LONG1 = b"\x8a"
LONG4 = b"\x8b"
BINBYTES = b"B"
SHORT_BINBYTES = b"C"
SHORT_BINUNICODE = b"\x8c"
BINUNICODE8 = b"\x8d"
BINBYTES8 = b"\x8e"
EMPTY_SET = b"\x8f"
ADDITEMS = b"\x90"
FROZENSET = b"\x91"
NEWOBJ_EX = b"\x92"
STACK_GLOBAL = b"\x93"
MEMOIZE = b"\x94"
FRAME = b"\x95"
BYTEARRAY8 = b"\x96"
NEXT_BUFFER = b"\x97"
READONLY_BUFFER = b"\x98"

class PickleError(Exception):
    pass


class PicklingError(PickleError):
    pass


class UnpicklingError(PickleError):
    pass


class _Stop(Exception):
    def __init__(self, value):
        self.value = value


def encode_long(x):
    if x == 0:
        return b""
    nbytes = (x.bit_length() >> 3) + 1
    result = x.to_bytes(nbytes, "little", signed=True)
    if x < 0 and nbytes > 1:
        if result[-1] == 0xFF and (result[-2] & 0x80) != 0:
            result = result[:-1]
    return result


def decode_long(data):
    return int.from_bytes(data, "little", signed=True)


class _Placeholder:
    pass


def _pack_name(s):
    return s.encode("utf-8")


class Pickler:
    def __init__(self, file, protocol=None, fix_imports=True, buffer_callback=None):
        if protocol is None:
            protocol = DEFAULT_PROTOCOL
        if protocol < 0:
            protocol = HIGHEST_PROTOCOL
        elif not 0 <= protocol <= HIGHEST_PROTOCOL:
            raise ValueError("pickle protocol must be <= %d" % HIGHEST_PROTOCOL)
        if not hasattr(file, "write"):
            raise TypeError("file must have a 'write' attribute")
        self._file = file
        self.proto = int(protocol)
        self.bin = protocol >= 1
        self.fast = 0
        self.fix_imports = fix_imports and protocol < 3
        self.memo = {}
        self._keep = []
        self._out = []
        self.dispatch_table = copyreg.dispatch_table

    def clear_memo(self):
        self.memo.clear()
        self._keep = []

    def _w(self, b):
        self._out.append(b)

    def dump(self, obj):
        self._out = []
        if self.proto >= 2:
            head = PROTO + bytes([self.proto])
        else:
            head = b""
        self.save(obj)
        self._w(STOP)
        body = b"".join(self._out)
        self._out = []
        if self.proto >= 4 and len(body) >= 4:
            body = FRAME + struct.pack("<Q", len(body)) + body
        self._file.write(head + body)

    def memoize(self, obj):
        idx = len(self.memo)
        self.memo[id(obj)] = (idx, obj)
        if self.proto >= 4:
            self._w(MEMOIZE)
        elif self.bin:
            if idx < 256:
                self._w(BINPUT + bytes([idx]))
            else:
                self._w(LONG_BINPUT + struct.pack("<I", idx))
        else:
            self._w(PUT + str(idx).encode("ascii") + b"\n")

    def _get(self, idx):
        if self.bin:
            if idx < 256:
                return BINGET + bytes([idx])
            return LONG_BINGET + struct.pack("<I", idx)
        return GET + str(idx).encode("ascii") + b"\n"

    def save(self, obj):
        if obj is None:
            self._w(NONE)
            return
        if obj is True:
            self._w(NEWTRUE if self.proto >= 2 else TRUE)
            return
        if obj is False:
            self._w(NEWFALSE if self.proto >= 2 else FALSE)
            return
        t = type(obj)
        if t is int:
            self.save_long(obj)
            return
        if t is float:
            self.save_float(obj)
            return
        x = self.memo.get(id(obj))
        if x is not None:
            self._w(self._get(x[0]))
            return
        if t is str:
            self.save_str(obj)
        elif t is bytes:
            self.save_bytes(obj)
        elif t is bytearray:
            self.save_bytearray(obj)
        elif t is tuple:
            self.save_tuple(obj)
        elif t is list:
            self.save_list(obj)
        elif t is dict:
            self.save_dict(obj)
        elif t is set:
            self.save_set(obj)
        elif t is frozenset:
            self.save_frozenset(obj)
        elif t is complex:
            self.save_reduce(complex, (obj.real, obj.imag), obj=obj)
        elif t is range:
            self.save_reduce(range, (obj.start, obj.stop, obj.step), obj=obj)
        elif t is slice:
            self.save_reduce(slice, (obj.start, obj.stop, obj.step), obj=obj)
        elif _native("pickle.is_class", obj):
            self.save_global(obj)
        elif _native("pickle.is_instance", obj) or _native("pickle.is_exception", obj):
            self.save_instance(obj)
        elif callable(obj):
            self.save_global(obj)
        else:
            raise TypeError("cannot pickle '" + t.__name__ + "' object")

    def save_long(self, obj):
        if self.bin:
            if obj >= 0:
                if obj <= 0xFF:
                    self._w(BININT1 + bytes([obj]))
                    return
                if obj <= 0xFFFF:
                    self._w(BININT2 + bytes([obj & 0xFF, obj >> 8]))
                    return
            if -0x80000000 <= obj <= 0x7FFFFFFF:
                self._w(BININT + struct.pack("<i", obj))
                return
        if self.proto >= 2:
            enc = encode_long(obj)
            n = len(enc)
            if n < 256:
                self._w(LONG1 + bytes([n]) + enc)
            else:
                self._w(LONG4 + struct.pack("<i", n) + enc)
            return
        if -0x80000000 <= obj <= 0x7FFFFFFF:
            self._w(INT + repr(obj).encode("ascii") + b"\n")
        else:
            self._w(LONG + repr(obj).encode("ascii") + b"L\n")

    def save_float(self, obj):
        if self.bin:
            self._w(BINFLOAT + struct.pack(">d", obj))
        else:
            self._w(FLOAT + repr(obj).encode("ascii") + b"\n")

    def save_str(self, obj):
        if self.bin:
            enc = obj.encode("utf-8", "surrogatepass")
            n = len(enc)
            if n <= 0xFF and self.proto >= 4:
                self._w(SHORT_BINUNICODE + bytes([n]) + enc)
            elif n > 0xFFFFFFFF and self.proto >= 4:
                self._w(BINUNICODE8 + struct.pack("<Q", n) + enc)
            else:
                self._w(BINUNICODE + struct.pack("<I", n) + enc)
        else:
            obj = obj.replace("\\", "\\u005c")
            obj = obj.replace("\0", "\\u0000")
            obj = obj.replace("\n", "\\u000a")
            obj = obj.replace("\r", "\\u000d")
            obj = obj.replace("\x1a", "\\u001a")
            self._w(UNICODE + obj.encode("raw-unicode-escape") + b"\n")
        self.memoize(obj)

    def save_bytes(self, obj):
        if self.proto < 3:
            if not obj:
                self.save_reduce(bytes, (), obj=obj)
            else:
                self.save_reduce(_codecs_encode, (str(obj, "latin1"), "latin1"), obj=obj)
            return
        n = len(obj)
        if n <= 0xFF:
            self._w(SHORT_BINBYTES + bytes([n]) + obj)
        elif n > 0xFFFFFFFF and self.proto >= 4:
            self._w(BINBYTES8 + struct.pack("<Q", n) + obj)
        else:
            self._w(BINBYTES + struct.pack("<I", n) + obj)
        self.memoize(obj)

    def save_bytearray(self, obj):
        if self.proto < 5:
            if not obj:
                self.save_reduce(bytearray, (), obj=obj)
            else:
                self.save_reduce(bytearray, (bytes(obj),), obj=obj)
            return
        self._w(BYTEARRAY8 + struct.pack("<Q", len(obj)) + bytes(obj))
        self.memoize(obj)

    def save_tuple(self, obj):
        if not obj:
            if self.proto:
                self._w(EMPTY_TUPLE)
            else:
                self._w(MARK + TUPLE)
            return
        n = len(obj)
        if n <= 3 and self.proto >= 2:
            for element in obj:
                self.save(element)
            if id(obj) in self.memo:
                self._w(POP * n + self._get(self.memo[id(obj)][0]))
            else:
                self._w([TUPLE1, TUPLE2, TUPLE3][n - 1])
                self.memoize(obj)
            return
        if self.proto:
            self._w(MARK)
            for element in obj:
                self.save(element)
        else:
            self._w(MARK)
            for element in obj:
                self.save(element)
        if id(obj) in self.memo:
            if self.bin:
                self._w(POP_MARK + self._get(self.memo[id(obj)][0]))
            else:
                self._w(POP * (n + 1) + self._get(self.memo[id(obj)][0]))
            return
        self._w(TUPLE)
        self.memoize(obj)

    def save_list(self, obj):
        if self.bin:
            self._w(EMPTY_LIST)
        else:
            self._w(MARK + LIST)
        self.memoize(obj)
        self._batch_appends(obj)

    def _batch_appends(self, items):
        if not self.bin:
            for x in items:
                self.save(x)
                self._w(APPEND)
            return
        i = 0
        n = len(items)
        while i < n:
            batch = items[i:i + 1000]
            i += 1000
            if len(batch) > 1:
                self._w(MARK)
                for x in batch:
                    self.save(x)
                self._w(APPENDS)
            else:
                self.save(batch[0])
                self._w(APPEND)

    def save_dict(self, obj):
        if self.bin:
            self._w(EMPTY_DICT)
        else:
            self._w(MARK + DICT)
        self.memoize(obj)
        self._batch_setitems(list(obj.items()))

    def _batch_setitems(self, items):
        if not self.bin:
            for k, v in items:
                self.save(k)
                self.save(v)
                self._w(SETITEM)
            return
        i = 0
        n = len(items)
        while i < n:
            batch = items[i:i + 1000]
            i += 1000
            if len(batch) > 1:
                self._w(MARK)
                for k, v in batch:
                    self.save(k)
                    self.save(v)
                self._w(SETITEMS)
            else:
                self.save(batch[0][0])
                self.save(batch[0][1])
                self._w(SETITEM)

    def save_set(self, obj):
        if self.proto < 4:
            self.save_reduce(set, (list(obj),), obj=obj)
            return
        self._w(EMPTY_SET)
        self.memoize(obj)
        items = list(obj)
        i = 0
        while i < len(items):
            batch = items[i:i + 1000]
            i += 1000
            self._w(MARK)
            for item in batch:
                self.save(item)
            self._w(ADDITEMS)

    def save_frozenset(self, obj):
        if self.proto < 4:
            self.save_reduce(frozenset, (list(obj),), obj=obj)
            return
        self._w(MARK)
        for item in obj:
            self.save(item)
        if id(obj) in self.memo:
            self._w(POP_MARK + self._get(self.memo[id(obj)][0]))
            return
        self._w(FROZENSET)
        self.memoize(obj)

    def _global_names(self, obj):
        gn = _native("pickle.global_name", obj)
        if gn is None:
            raise PicklingError("Can't pickle " + repr(obj) + ": it's not found as a global")
        module = gn[0]
        name = gn[1]
        if module == "pickle" and name == "_codecs_encode":
            return "_codecs", "encode"
        return module, name

    def save_global(self, obj, name=None):
        module, name = self._global_names(obj)
        if self.proto >= 4:
            self.save(module)
            self.save(name)
            self._w(STACK_GLOBAL)
        else:
            if self.proto >= 3:
                pass
            elif self.fix_imports:
                key = (module, name)
                if key in _compat_pickle.REVERSE_NAME_MAPPING:
                    module, name = _compat_pickle.REVERSE_NAME_MAPPING[key]
                elif module in _compat_pickle.REVERSE_IMPORT_MAPPING:
                    module = _compat_pickle.REVERSE_IMPORT_MAPPING[module]
            self._w(GLOBAL + module.encode("utf-8") + b"\n" + name.encode("utf-8") + b"\n")
        self.memoize(obj)

    def _reduce_value(self, obj):
        cls = type(obj)
        if cls in copyreg.dispatch_table:
            return copyreg.dispatch_table[cls](obj)
        if hasattr(obj, "__reduce_ex__"):
            return obj.__reduce_ex__(self.proto)
        if hasattr(obj, "__reduce__"):
            return obj.__reduce__()
        return self._default_reduce(obj)

    def _state(self, obj):
        if hasattr(obj, "__getstate__"):
            return obj.__getstate__()
        d = vars(obj)
        if not d:
            return None
        return dict(d)

    def _default_reduce(self, obj):
        cls = type(obj)
        if _native("pickle.is_exception", obj):
            ed = {}
            try:
                for k, v in vars(obj).items():
                    ed[k] = v
            except TypeError:
                pass
            if ed:
                return (cls, obj.args, ed)
            return (cls, obj.args)
        state = self._state(obj)
        if self.proto >= 2:
            args = ()
            kwargs = None
            if hasattr(obj, "__getnewargs_ex__"):
                ga = obj.__getnewargs_ex__()
                args = ga[0]
                kwargs = ga[1]
            elif hasattr(obj, "__getnewargs__"):
                args = obj.__getnewargs__()
            if kwargs:
                return (copyreg.__newobj_ex__, (cls, args, kwargs), state)
            return (copyreg.__newobj__, (cls,) + tuple(args), state)
        return (copyreg._reconstructor, (cls, object, None), state)

    def save_instance(self, obj):
        rv = self._reduce_value(obj)
        if isinstance(rv, str):
            self.save_global(obj)
            return
        if not isinstance(rv, tuple):
            raise PicklingError("__reduce__ must return a string or tuple")
        n = len(rv)
        if n < 2 or n > 6:
            raise PicklingError("tuple returned by __reduce__ must contain 2 through 6 elements")
        func = rv[0]
        args = rv[1]
        state = rv[2] if n > 2 else None
        listitems = rv[3] if n > 3 else None
        dictitems = rv[4] if n > 4 else None
        if not callable(func):
            raise PicklingError("first item of the tuple returned by __reduce__ must be callable")
        if not isinstance(args, tuple):
            raise PicklingError("second item of the tuple returned by __reduce__ must be a tuple")
        self.save_reduce(func, args, state, listitems, dictitems, obj=obj)

    def save_reduce(self, func, args, state=None, listitems=None, dictitems=None, state_setter=None, obj=None):
        fname = _native("pickle.func_name", func)
        if self.proto >= 2 and fname == "__newobj_ex__":
            cls = args[0]
            cargs = args[1]
            kwargs = args[2]
            if self.proto >= 4:
                self.save(cls)
                self.save(cargs)
                self.save(kwargs)
                self._w(NEWOBJ_EX)
            else:
                self.save(copyreg.__newobj_ex__)
                self.save((cls, cargs, kwargs))
                self._w(REDUCE)
        elif self.proto >= 2 and fname == "__newobj__":
            if not args:
                raise PicklingError("__newobj__ arguments=" + repr(args))
            cls = args[0]
            if obj is not None and cls is not type(obj):
                raise PicklingError("args[0] from __newobj__ args has the wrong class")
            rest = args[1:]
            self.save(cls)
            self.save(rest)
            self._w(NEWOBJ)
        else:
            self.save(func)
            self.save(args)
            self._w(REDUCE)
        if obj is not None:
            x = self.memo.get(id(obj))
            if x is not None:
                self._w(POP + self._get(x[0]))
            else:
                self.memoize(obj)
        if listitems is not None:
            self._batch_appends(list(listitems))
        if dictitems is not None:
            self._batch_setitems(list(dictitems))
        if state is not None:
            self.save(state)
            self._w(BUILD)


def _codecs_encode(s, enc):
    return s.encode(enc)


class _Reader:
    def __init__(self, file, data):
        self.file = file
        self.data = data
        self.pos = 0

    def read(self, n):
        if self.file is None:
            if self.pos + n > len(self.data):
                self.pos = len(self.data)
                raise UnpicklingError("pickle data was truncated")
            r = self.data[self.pos:self.pos + n]
            self.pos += n
            return r
        r = self.file.read(n)
        if len(r) < n:
            raise UnpicklingError("pickle data was truncated")
        return r

    def readline(self):
        if self.file is None:
            i = self.data.find(b"\n", self.pos)
            if i < 0:
                raise UnpicklingError("pickle data was truncated")
            r = self.data[self.pos:i + 1]
            self.pos = i + 1
            return r
        r = self.file.readline()
        if not r.endswith(b"\n"):
            raise UnpicklingError("pickle data was truncated")
        return r


class Unpickler:
    def __init__(self, file, fix_imports=True, encoding="ASCII", errors="strict", buffers=None):
        self._file = file
        self.fix_imports = fix_imports
        self.encoding = encoding
        self.errors = errors
        self.memo = {}
        self.stack = []
        self.marks = []
        self.proto = 0

    def persistent_load(self, pid):
        raise UnpicklingError("unsupported persistent id encountered")

    def find_class(self, module, name):
        if module == "_codecs" and name == "encode":
            return _codecs_encode
        if self.proto < 3 and self.fix_imports:
            key = (module, name)
            if key in _compat_pickle.NAME_MAPPING:
                module, name = _compat_pickle.NAME_MAPPING[key]
            elif module in _compat_pickle.IMPORT_MAPPING:
                module = _compat_pickle.IMPORT_MAPPING[module]
        return _native("pickle.find_global", module, name)

    def load(self):
        if not hasattr(self._file, "read"):
            raise TypeError("file must have 'read' and 'readline' attributes")
        r = _Reader(self._file, b"")
        self._r = r
        return self._run(r)

    def _load_bytes(self, data):
        r = _Reader(None, data)
        self._r = r
        return self._run(r)

    def _marker(self):
        if not self.marks:
            raise UnpicklingError("could not find MARK")
        m = self.marks.pop()
        items = self.stack[m:]
        del self.stack[m:]
        return items

    def _pop(self):
        if self.marks and len(self.stack) == self.marks[-1]:
            self.marks.pop()
        if not self.stack:
            raise UnpicklingError("unpickling stack underflow")
        return self.stack.pop()

    def _run(self, r):
        self.stack = []
        self.marks = []
        self.memo = {}
        self.proto = 0
        while True:
            if r.file is None:
                if r.pos >= len(r.data):
                    raise EOFError("Ran out of input")
            else:
                key = r.file.read(1)
                if not key:
                    raise EOFError("Ran out of input")
                r.data = key
                r.pos = 0
            if r.file is None:
                key = r.read(1)
            try:
                done = self._step(key[0], r)
            except IndexError:
                raise UnpicklingError("unpickling stack underflow")
            if done:
                return self.stack.pop()

    def _step(self, op, r):
        stack = self.stack
        c = bytes([op])
        if c == STOP:
            if not stack:
                raise UnpicklingError("unpickling stack underflow")
            return True
        elif c == PROTO:
            p = r.read(1)[0]
            if p > HIGHEST_PROTOCOL:
                raise ValueError("unsupported pickle protocol: %d" % p)
            self.proto = p
        elif c == FRAME:
            r.read(8)
        elif c == NONE:
            stack.append(None)
        elif c == NEWTRUE:
            stack.append(True)
        elif c == NEWFALSE:
            stack.append(False)
        elif c == INT:
            line = r.readline()[:-1]
            if line == b"00":
                stack.append(False)
            elif line == b"01":
                stack.append(True)
            else:
                stack.append(int(line))
        elif c == LONG:
            line = r.readline()[:-1]
            if line.endswith(b"L"):
                line = line[:-1]
            stack.append(int(line))
        elif c == BININT:
            stack.append(struct.unpack("<i", r.read(4))[0])
        elif c == BININT1:
            stack.append(r.read(1)[0])
        elif c == BININT2:
            stack.append(struct.unpack("<H", r.read(2))[0])
        elif c == LONG1:
            n = r.read(1)[0]
            stack.append(decode_long(r.read(n)))
        elif c == LONG4:
            n = struct.unpack("<i", r.read(4))[0]
            stack.append(decode_long(r.read(n)))
        elif c == FLOAT:
            stack.append(float(r.readline()[:-1]))
        elif c == BINFLOAT:
            stack.append(struct.unpack(">d", r.read(8))[0])
        elif c == STRING:
            line = r.readline()[:-1]
            if len(line) >= 2 and line[0] == line[-1] and line[0] in b"\"'":
                line = line[1:-1]
            else:
                raise UnpicklingError("the STRING opcode argument must be quoted")
            stack.append(line.decode("latin-1") if self.encoding == "bytes" else line.decode(self.encoding, self.errors))
        elif c == BINSTRING:
            n = struct.unpack("<i", r.read(4))[0]
            stack.append(r.read(n).decode(self.encoding, self.errors))
        elif c == SHORT_BINSTRING:
            n = r.read(1)[0]
            stack.append(r.read(n).decode(self.encoding, self.errors))
        elif c == BINBYTES:
            n = struct.unpack("<I", r.read(4))[0]
            stack.append(r.read(n))
        elif c == SHORT_BINBYTES:
            n = r.read(1)[0]
            stack.append(r.read(n))
        elif c == BINBYTES8:
            n = struct.unpack("<Q", r.read(8))[0]
            stack.append(r.read(n))
        elif c == BYTEARRAY8:
            n = struct.unpack("<Q", r.read(8))[0]
            stack.append(bytearray(r.read(n)))
        elif c == UNICODE:
            stack.append(r.readline()[:-1].decode("raw-unicode-escape"))
        elif c == BINUNICODE:
            n = struct.unpack("<I", r.read(4))[0]
            stack.append(r.read(n).decode("utf-8", "surrogatepass"))
        elif c == SHORT_BINUNICODE:
            n = r.read(1)[0]
            stack.append(r.read(n).decode("utf-8", "surrogatepass"))
        elif c == BINUNICODE8:
            n = struct.unpack("<Q", r.read(8))[0]
            stack.append(r.read(n).decode("utf-8", "surrogatepass"))
        elif c == EMPTY_TUPLE:
            stack.append(())
        elif c == TUPLE:
            items = self._marker()
            stack.append(tuple(items))
        elif c == TUPLE1:
            stack[-1] = (stack[-1],)
        elif c == TUPLE2:
            stack[-2:] = [(stack[-2], stack[-1])]
        elif c == TUPLE3:
            stack[-3:] = [(stack[-3], stack[-2], stack[-1])]
        elif c == EMPTY_LIST:
            stack.append([])
        elif c == EMPTY_DICT:
            stack.append({})
        elif c == EMPTY_SET:
            stack.append(set())
        elif c == MARK:
            self.marks.append(len(stack))
        elif c == POP:
            self._pop()
        elif c == POP_MARK:
            self._marker()
        elif c == DUP:
            stack.append(stack[-1])
        elif c == GET:
            stack.append(self.memo[int(r.readline()[:-1])])
        elif c == BINGET:
            stack.append(self._memo_get(r.read(1)[0]))
        elif c == LONG_BINGET:
            stack.append(self._memo_get(struct.unpack("<I", r.read(4))[0]))
        elif c == PUT:
            i = int(r.readline()[:-1])
            if i < 0:
                raise ValueError("negative PUT argument")
            self.memo[i] = stack[-1]
        elif c == BINPUT:
            self.memo[r.read(1)[0]] = stack[-1]
        elif c == LONG_BINPUT:
            self.memo[struct.unpack("<I", r.read(4))[0]] = stack[-1]
        elif c == MEMOIZE:
            self.memo[len(self.memo)] = stack[-1]
        elif c == APPEND:
            v = self._pop()
            lst = stack[-1]
            lst.append(v)
        elif c == APPENDS:
            items = self._marker()
            stack[-1].extend(items)
        elif c == LIST:
            items = self._marker()
            stack.append(list(items))
        elif c == SETITEM:
            v = self._pop()
            k = self._pop()
            stack[-1][k] = v
        elif c == SETITEMS:
            items = self._marker()
            dst = stack[-1]
            i = 0
            while i + 1 < len(items):
                dst[items[i]] = items[i + 1]
                i += 2
        elif c == DICT:
            items = self._marker()
            nd = {}
            i = 0
            while i + 1 < len(items):
                nd[items[i]] = items[i + 1]
                i += 2
            stack.append(nd)
        elif c == ADDITEMS:
            items = self._marker()
            s = stack[-1]
            for x in items:
                s.add(x)
        elif c == FROZENSET:
            items = self._marker()
            stack.append(frozenset(items))
        elif c == GLOBAL:
            module = r.readline()[:-1].decode("utf-8")
            name = r.readline()[:-1].decode("utf-8")
            stack.append(self.find_class(module, name))
        elif c == STACK_GLOBAL:
            name = self._pop()
            module = self._pop()
            if not isinstance(module, str) or not isinstance(name, str):
                raise UnpicklingError("STACK_GLOBAL requires str")
            stack.append(self.find_class(module, name))
        elif c == REDUCE:
            args = self._pop()
            func = stack[-1]
            stack[-1] = func(*args)
        elif c == NEWOBJ:
            args = self._pop()
            cls = stack[-1]
            stack[-1] = _native("pickle.newobj", cls, args)
        elif c == NEWOBJ_EX:
            kwargs = self._pop()
            args = self._pop()
            cls = stack[-1]
            if kwargs:
                stack[-1] = cls(*args, **kwargs)
            else:
                stack[-1] = _native("pickle.newobj", cls, args)
        elif c == BUILD:
            state = self._pop()
            inst = stack[-1]
            if hasattr(inst, "__setstate__"):
                inst.__setstate__(state)
            else:
                slotstate = None
                if isinstance(state, tuple) and len(state) == 2:
                    state, slotstate = state
                if state:
                    for k, v in state.items():
                        setattr(inst, k, v)
                if slotstate:
                    for k, v in slotstate.items():
                        setattr(inst, k, v)
        elif c == OBJ:
            items = self._marker()
            cls = items[0]
            stack.append(cls(*items[1:]))
        elif c == INST:
            module = r.readline()[:-1].decode("utf-8")
            name = r.readline()[:-1].decode("utf-8")
            items = self._marker()
            cls = self.find_class(module, name)
            stack.append(cls(*items))
        elif c == PERSID:
            pid = r.readline()[:-1].decode("ascii")
            stack.append(self.persistent_load(pid))
        elif c == BINPERSID:
            pid = self._pop()
            stack.append(self.persistent_load(pid))
        elif c == EXT1 or c == EXT2 or c == EXT4:
            n = 1 if c == EXT1 else (2 if c == EXT2 else 4)
            code = int.from_bytes(r.read(n), "little")
            key = copyreg._inverted_registry.get(code)
            if key is None:
                raise ValueError("unregistered extension code %d" % code)
            stack.append(self.find_class(key[0], key[1]))
        else:
            ch = chr(op)
            if 32 <= op < 127:
                raise UnpicklingError("invalid load key, '%s'." % ch)
            raise UnpicklingError("invalid load key, '\\x%02x'." % op)
        return False

    def _memo_get(self, i):
        if i not in self.memo:
            raise UnpicklingError("Memo value not found at index %d" % i)
        return self.memo[i]


class _Sink:
    def __init__(self):
        self.parts = []

    def write(self, b):
        self.parts.append(b)


def dump(obj, file, protocol=None, fix_imports=True, buffer_callback=None):
    Pickler(file, protocol, fix_imports).dump(obj)


def dumps(obj, protocol=None, fix_imports=True, buffer_callback=None):
    sink = _Sink()
    Pickler(sink, protocol, fix_imports).dump(obj)
    return b"".join(sink.parts)


def load(file, fix_imports=True, encoding="ASCII", errors="strict", buffers=None):
    return Unpickler(file, fix_imports, encoding, errors).load()


def loads(data, fix_imports=True, encoding="ASCII", errors="strict", buffers=None):
    if isinstance(data, str):
        raise TypeError("a bytes-like object is required, not 'str'")
    return Unpickler(None, fix_imports, encoding, errors)._load_bytes(bytes(data))


_Pickler = Pickler
_Unpickler = Unpickler
