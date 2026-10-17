"""_tokenize: the tokenizer behind the tokenize module, written in Go
(runtime/tokenizer.go). It gives what CPython's C tokenizer gives with
extra_tokens=True."""


class TokenizerIter:
    """The tokens of the lines that readline returns, as tuples (type,
    string, start, end, line). The lines are read when the iteration starts;
    a tokenizing error is raised after the tokens before it."""

    def __init__(self, readline, *, extra_tokens=False, encoding=None):
        if not extra_tokens:
            raise NotImplementedError("py2go's tokenizer gives the tokens of the tokenize module only (extra_tokens=True)")
        self._readline = readline
        self._encoding = encoding
        self._tokens = None
        self._error = None
        self._pos = 0

    def __iter__(self):
        return self

    def _read(self):
        parts = []
        while True:
            try:
                line = self._readline()
            except StopIteration:
                break
            if not line:
                break
            if self._encoding is not None:
                if isinstance(line, bytes):
                    line = line.decode(self._encoding)
            elif not isinstance(line, str):
                raise TypeError("readline() returned a non-string object")
            parts.append(line)
        tokens, error = _native("tokenize.tokens", "".join(parts))
        self._tokens = tokens
        self._error = error

    def __next__(self):
        if self._tokens is None:
            self._read()
        if self._pos < len(self._tokens):
            tok = self._tokens[self._pos]
            self._pos += 1
            return tok
        if self._error is not None:
            kind, msg, lineno, offset = self._error
            self._error = None
            if kind == "IndentationError":
                raise IndentationError(msg, ("<string>", lineno, offset, None, None, None))
            if kind == "TabError":
                raise TabError(msg, ("<string>", lineno, offset, None, None, None))
            raise SyntaxError(msg, ("<string>", lineno, offset, None, None, None))
        raise StopIteration
