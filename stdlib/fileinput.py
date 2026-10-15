"""fileinput: iterate over lines from multiple input streams."""
import io
import os
import sys

__all__ = ["input", "close", "nextfile", "filename", "lineno", "filelineno", "fileno", "isfirstline", "isstdin",
           "FileInput", "hook_compressed", "hook_encoded"]

_state = None


def input(files=None, inplace=False, backup="", *, mode="r", openhook=None, encoding=None, errors=None):
    global _state
    if _state is not None and _state._file:
        raise RuntimeError("input() already active")
    _state = FileInput(files, inplace, backup, mode=mode, openhook=openhook, encoding=encoding, errors=errors)
    return _state


def close():
    global _state
    state = _state
    _state = None
    if state:
        state.close()


def nextfile():
    return _state.nextfile()


def filename():
    return _state.filename()


def lineno():
    return _state.lineno()


def filelineno():
    return _state.filelineno()


def fileno():
    return _state.fileno()


def isfirstline():
    return _state.isfirstline()


def isstdin():
    return _state.isstdin()


class FileInput:
    def __init__(self, files=None, inplace=False, backup="", *, mode="r", openhook=None, encoding=None, errors=None):
        if isinstance(files, str):
            files = (files,)
        elif isinstance(files, os.PathLike):
            files = (os.fspath(files),)
        else:
            if files is None:
                files = sys.argv[1:]
            if not files:
                files = ("-",)
            else:
                files = tuple(files)
        self._files = files
        self._inplace = inplace
        self._backup = backup
        self._savestdout = None
        self._output = None
        self._filename = None
        self._startlineno = 0
        self._filelineno = 0
        self._file = None
        self._isstdin = False
        self._backupfilename = None
        self._encoding = encoding
        self._errors = errors
        self._lines = []
        self._pos = 0
        self._fileidx = 0
        if mode not in ("r", "rb"):
            raise ValueError("FileInput opening mode must be 'r' or 'rb'")
        self._mode = mode
        self._write_mode = "w" if "b" not in mode else "wb"
        self._openhook = openhook
        if inplace and openhook:
            raise ValueError("FileInput cannot use an opening hook in inplace mode")


    def close(self):
        try:
            self.nextfile()
        finally:
            self._files = ()

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()

    def __iter__(self):
        return self

    def __next__(self):
        while True:
            line = self._readline()
            if line:
                self._filelineno += 1
                return line
            if not self._file:
                raise StopIteration
            self.nextfile()

    def nextfile(self):
        self._file = None
        self._lines = []
        self._pos = 0
        self._startlineno = self._startlineno + self._filelineno if False else self._startlineno

    def _readline(self):
        if self._pos < len(self._lines):
            line = self._lines[self._pos]
            self._pos += 1
            return line
        if not self._files:
            return "" if "b" not in self._mode else b""
        if self._fileidx >= len(self._files):
            return "" if "b" not in self._mode else b""
        self._filename = self._files[self._fileidx]
        self._fileidx += 1
        self._startlineno = self._startlineno + self._filelineno
        self._filelineno = 0
        self._isstdin = False
        if self._filename == "-":
            self._filename = "<stdin>"
            self._isstdin = True
            data = sys.stdin.read()
            lines = data.splitlines(keepends=True)
        else:
            if self._openhook:
                f = self._openhook(self._filename, self._mode)
            elif "b" in self._mode:
                f = io.open(self._filename, self._mode)
            else:
                f = io.open(self._filename, self._mode, encoding=self._encoding, errors=self._errors)
            lines = f.readlines()
            f.close()
        self._file = True
        self._lines = lines
        self._pos = 0
        return self._readline()

    def readline(self):
        while True:
            line = self._readline()
            if line:
                self._filelineno += 1
                return line
            if not self._file:
                return line
            self.nextfile()

    def filename(self):
        return self._filename

    def lineno(self):
        return self._startlineno + self._filelineno

    def filelineno(self):
        return self._filelineno

    def fileno(self):
        return -1

    def isfirstline(self):
        return self._filelineno == 1

    def isstdin(self):
        return self._isstdin


def hook_compressed(filename, mode, *, encoding=None, errors=None):
    ext = os.path.splitext(filename)[1]
    if ext == ".gz":
        import gzip
        return gzip.open(filename, mode)
    return io.open(filename, mode, encoding=encoding, errors=errors) if "b" not in mode else io.open(filename, mode)


def hook_encoded(encoding, errors=None):
    def openhook(filename, mode):
        return io.open(filename, mode, encoding=encoding, errors=errors)
    return openhook
