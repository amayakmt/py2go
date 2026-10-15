"""logging: a logging facility for applications (a port of CPython's logging)."""
import sys
import os
import time
import io
import threading
import warnings
from string import Template

__all__ = ["BASIC_FORMAT", "BufferingFormatter", "CRITICAL", "DEBUG", "ERROR",
           "FATAL", "FileHandler", "Filter", "Formatter", "Handler", "INFO",
           "LogRecord", "Logger", "LoggerAdapter", "NOTSET", "NullHandler",
           "StreamHandler", "WARN", "WARNING", "addLevelName", "basicConfig",
           "captureWarnings", "critical", "debug", "disable", "error",
           "exception", "fatal", "getLevelName", "getLogger", "getLoggerClass",
           "info", "log", "makeLogRecord", "setLoggerClass", "shutdown",
           "warn", "warning", "getLogRecordFactory", "setLogRecordFactory",
           "lastResort", "raiseExceptions", "getLevelNamesMapping",
           "getHandlerByName", "getHandlerNames"]

_startTime = time.time()


def _obj(x: object) -> object:
    return x


raiseExceptions = True
logThreads = True
logMultiprocessing = True
logProcesses = True
logAsyncioTasks = True

CRITICAL = 50
FATAL = CRITICAL
ERROR = 40
WARNING = 30
WARN = WARNING
INFO = 20
DEBUG = 10
NOTSET = 0

_levelToName = _obj({
    CRITICAL: "CRITICAL",
    ERROR: "ERROR",
    WARNING: "WARNING",
    INFO: "INFO",
    DEBUG: "DEBUG",
    NOTSET: "NOTSET",
})
_nameToLevel = _obj({
    "CRITICAL": CRITICAL,
    "FATAL": FATAL,
    "ERROR": ERROR,
    "WARN": WARNING,
    "WARNING": WARNING,
    "INFO": INFO,
    "DEBUG": DEBUG,
    "NOTSET": NOTSET,
})


def getLevelNamesMapping():
    return dict(_nameToLevel)


def getLevelName(level):
    result = _levelToName.get(level)
    if result is not None:
        return result
    result = _nameToLevel.get(level)
    if result is not None:
        return result
    return "Level %s" % level


def addLevelName(level, levelName):
    _acquireLock()
    try:
        _levelToName[level] = levelName
        _nameToLevel[levelName] = level
    finally:
        _releaseLock()


def _checkLevel(level):
    if isinstance(level, int):
        rv = level
    elif str(level) == level:
        if level not in _nameToLevel:
            raise ValueError("Unknown level: %r" % level)
        rv = _nameToLevel[level]
    else:
        raise TypeError("Level not an integer or a valid string: %r" % (level,))
    return rv


_lock = threading.RLock()


def _acquireLock():
    if _lock:
        _lock.acquire()


def _releaseLock():
    if _lock:
        _lock.release()


def _prepareFork():
    pass


class LogRecord:
    def __init__(self, name, level, pathname, lineno, msg, args, exc_info, func=None, sinfo=None, **kwargs):
        ct = time.time()
        self.name = name
        self.msg = msg
        if args and len(args) == 1 and isinstance(args[0], dict) and args[0]:
            args = args[0]
        self.args = args
        self.levelname = getLevelName(level)
        self.levelno = level
        self.pathname = pathname
        try:
            self.filename = os.path.basename(pathname)
            self.module = os.path.splitext(self.filename)[0]
        except (TypeError, ValueError, AttributeError):
            self.filename = pathname
            self.module = "Unknown module"
        self.exc_info = exc_info
        self.exc_text = None
        self.stack_info = sinfo
        self.lineno = lineno
        self.funcName = func
        self.created = ct
        self.msecs = int((ct - int(ct)) * 1000) + 0.0
        self.relativeCreated = (self.created - _startTime) * 1000
        if logThreads:
            self.thread = threading.get_ident()
            self.threadName = threading.current_thread().name
        else:
            self.thread = None
            self.threadName = None
        if not logMultiprocessing:
            self.processName = None
        else:
            self.processName = "MainProcess"
        if logProcesses:
            self.process = os.getpid()
        else:
            self.process = None
        self.taskName = None

    def __repr__(self):
        return '<LogRecord: %s, %s, %s, %s, "%s">' % (self.name, self.levelno, self.pathname, self.lineno, self.msg)

    def getMessage(self):
        msg = str(self.msg)
        if self.args:
            msg = msg % self.args
        return msg


_logRecordFactory = LogRecord


def setLogRecordFactory(factory):
    global _logRecordFactory
    _logRecordFactory = factory


def getLogRecordFactory():
    return _logRecordFactory


def makeLogRecord(dict):
    rv = _logRecordFactory(None, None, "", 0, "", (), None, None)
    for k in dict:
        setattr(rv, k, dict[k])
    return rv


class PercentStyle:
    default_format = "%(message)s"
    asctime_format = "%(asctime)s"
    asctime_search = "%(asctime)"
    validation_pattern = None

    def __init__(self, fmt, *, defaults=None):
        if fmt:
            self._fmt = fmt
        else:
            self._fmt = self.default_format
        self._defaults = defaults

    def usesTime(self):
        return self._fmt.find(self.asctime_search) >= 0

    def validate(self):
        import re
        if not re.search(r"%\(\w+\)[#0+ -]*(\*|\d+)?(\.(\*|\d+))?[diouxefgcrsa%]", self._fmt, re.I):
            raise ValueError("Invalid format '%s' for '%s' style" % (self._fmt, "%"))

    def _format(self, record):
        if self._defaults is not None:
            values = dict(self._defaults)
            values.update(record.__dict__)
        else:
            values = record.__dict__
        return self._fmt % values

    def format(self, record):
        try:
            return self._format(record)
        except KeyError as e:
            raise ValueError("Formatting field not found in record: %s" % str(e))


class StrFormatStyle(PercentStyle):
    default_format = "{message}"
    asctime_format = "{asctime}"
    asctime_search = "{asctime"

    def _format(self, record):
        if self._defaults is not None:
            values = dict(self._defaults)
            values.update(record.__dict__)
        else:
            values = record.__dict__
        return self._fmt.format(**values)

    def validate(self):
        try:
            self._fmt.format(**{"message": "", "asctime": "", "levelname": "", "name": ""})
        except KeyError:
            pass
        except (ValueError, IndexError) as e:
            raise ValueError("invalid format: %s" % str(e))


class StringTemplateStyle(PercentStyle):
    default_format = "${message}"
    asctime_format = "${asctime}"
    asctime_search = "${asctime}"

    def __init__(self, fmt, *, defaults=None):
        self._fmt = fmt or self.default_format
        self._tpl = Template(self._fmt)
        self._defaults = defaults

    def usesTime(self):
        fmt = self._fmt
        return fmt.find("$asctime") >= 0 or fmt.find(self.asctime_search) >= 0

    def validate(self):
        pass

    def _format(self, record):
        if self._defaults is not None:
            values = dict(self._defaults)
            values.update(record.__dict__)
        else:
            values = record.__dict__
        return self._tpl.substitute(**values)


BASIC_FORMAT = "%(levelname)s:%(name)s:%(message)s"

_STYLES = {
    "%": (PercentStyle, BASIC_FORMAT),
    "{": (StrFormatStyle, "{levelname}:{name}:{message}"),
    "$": (StringTemplateStyle, "${levelname}:${name}:${message}"),
}


class Formatter:
    converter = time.localtime

    def __init__(self, fmt=None, datefmt=None, style="%", validate=True, *, defaults=None):
        if style not in _STYLES:
            raise ValueError("Style must be one of: %s" % ",".join(_STYLES.keys()))
        self._style = _STYLES[style][0](fmt, defaults=defaults)
        if validate:
            self._style.validate()
        self._fmt = self._style._fmt
        self.datefmt = datefmt

    default_time_format = "%Y-%m-%d %H:%M:%S"
    default_msec_format = "%s,%03d"

    def formatTime(self, record, datefmt=None):
        ct = self.converter(record.created)
        if datefmt:
            s = time.strftime(datefmt, ct)
        else:
            s = time.strftime(self.default_time_format, ct)
            if self.default_msec_format:
                s = self.default_msec_format % (s, record.msecs)
        return s

    def formatException(self, ei):
        return _format_exception(ei)

    def usesTime(self):
        return self._style.usesTime()

    def formatMessage(self, record):
        return self._style.format(record)

    def formatStack(self, stack_info):
        return stack_info

    def format(self, record):
        record.message = record.getMessage()
        if self.usesTime():
            record.asctime = self.formatTime(record, self.datefmt)
        s = self.formatMessage(record)
        if record.exc_info:
            if not record.exc_text:
                record.exc_text = self.formatException(record.exc_info)
        if record.exc_text:
            if s[-1:] != "\n":
                s = s + "\n"
            s = s + record.exc_text
        if record.stack_info:
            if s[-1:] != "\n":
                s = s + "\n"
            s = s + self.formatStack(record.stack_info)
        return s


def _format_exception(ei):
    exc = ei[1]
    if exc is None:
        return ""
    name = type(exc).__name__
    text = str(exc)
    if text:
        line = name + ": " + text
    else:
        line = name
    return line


_defaultFormatter = Formatter()


class BufferingFormatter:
    def __init__(self, linefmt=None):
        if linefmt:
            self.linefmt = linefmt
        else:
            self.linefmt = _defaultFormatter

    def formatHeader(self, records):
        return ""

    def formatFooter(self, records):
        return ""

    def format(self, records):
        rv = ""
        if len(records) > 0:
            rv = rv + self.formatHeader(records)
            for record in records:
                rv = rv + self.linefmt.format(record)
            rv = rv + self.formatFooter(records)
        return rv


class Filter:
    def __init__(self, name=""):
        self.name = name
        self.nlen = len(name)

    def filter(self, record):
        if self.nlen == 0:
            return True
        elif self.name == record.name:
            return True
        elif record.name.find(self.name, 0, self.nlen) != 0:
            return False
        return record.name[self.nlen] == "."


class Filterer:
    def __init__(self):
        self.filters = []

    def addFilter(self, filter):
        if not (filter in self.filters):
            self.filters.append(filter)

    def removeFilter(self, filter):
        if filter in self.filters:
            self.filters.remove(filter)

    def filter(self, record):
        for f in self.filters:
            if hasattr(f, "filter"):
                result = f.filter(record)
            else:
                result = f(record)
            if not result:
                return False
            if isinstance(result, LogRecord):
                record = result
        return record


_handlers = {}
_handlerList = []


def _removeHandlerRef(wr):
    pass


def _addHandlerRef(handler):
    _acquireLock()
    try:
        _handlerList.append(handler)
    finally:
        _releaseLock()


def getHandlerByName(name):
    return _handlers.get(name)


def getHandlerNames():
    return frozenset(_handlers)


class Handler(Filterer):
    def __init__(self, level=NOTSET):
        Filterer.__init__(self)
        self._name = None
        self.level = _checkLevel(level)
        self.formatter = None
        self._closed = False
        _addHandlerRef(self)
        self.createLock()

    def get_name(self):
        return self._name

    def set_name(self, name):
        _acquireLock()
        try:
            if self._name in _handlers:
                del _handlers[self._name]
            self._name = name
            if name:
                _handlers[name] = self
        finally:
            _releaseLock()

    @property
    def name(self):
        return self._name

    @name.setter
    def name(self, name):
        self.set_name(name)

    def createLock(self):
        self.lock = threading.RLock()

    def acquire(self):
        if self.lock:
            self.lock.acquire()

    def release(self):
        if self.lock:
            self.lock.release()

    def setLevel(self, level):
        self.level = _checkLevel(level)

    def format(self, record):
        if self.formatter:
            fmt = self.formatter
        else:
            fmt = _defaultFormatter
        return fmt.format(record)

    def emit(self, record):
        raise NotImplementedError("emit must be implemented by Handler subclasses")

    def handle(self, record):
        rv = self.filter(record)
        if isinstance(rv, LogRecord):
            record = rv
        if rv:
            self.acquire()
            try:
                self.emit(record)
            finally:
                self.release()
        return rv

    def setFormatter(self, fmt):
        self.formatter = fmt

    def flush(self):
        pass

    def close(self):
        _acquireLock()
        try:
            self._closed = True
            if self._name and self._name in _handlers:
                del _handlers[self._name]
        finally:
            _releaseLock()

    def handleError(self, record):
        if raiseExceptions and sys.stderr:
            t, v, tb = sys.exc_info()
            sys.stderr.write("--- Logging error ---\n")
            if v is not None:
                sys.stderr.write(type(v).__name__ + ": " + str(v) + "\n")
            sys.stderr.write("Call stack:\n")
            try:
                sys.stderr.write("Message: %r\nArguments: %s\n" % (record.msg, record.args))
            except RecursionError:
                raise
            except Exception:
                sys.stderr.write("Unable to print the message and arguments - possible formatting error.\nUse the traceback above to help find the error.\n")

    def __repr__(self):
        level = getLevelName(self.level)
        return "<%s (%s)>" % (self.__class__.__name__, level)


class StreamHandler(Handler):
    terminator = "\n"

    def __init__(self, stream=None):
        Handler.__init__(self)
        if stream is None:
            stream = sys.stderr
        self.stream = stream

    def flush(self):
        self.acquire()
        try:
            if self.stream and hasattr(self.stream, "flush"):
                self.stream.flush()
        finally:
            self.release()

    def emit(self, record):
        try:
            msg = self.format(record)
            stream = self.stream
            stream.write(msg + self.terminator)
            self.flush()
        except RecursionError:
            raise
        except Exception:
            self.handleError(record)

    def setStream(self, stream):
        if stream is self.stream:
            result = None
        else:
            result = self.stream
            self.acquire()
            try:
                self.flush()
                self.stream = stream
            finally:
                self.release()
        return result

    def __repr__(self):
        level = getLevelName(self.level)
        name = getattr(self.stream, "name", "")
        name = str(name)
        if name:
            name += " "
        return "<%s %s(%s)>" % (self.__class__.__name__, name, level)


class FileHandler(StreamHandler):
    def __init__(self, filename, mode="a", encoding=None, delay=False, errors=None):
        filename = os.fspath(filename)
        self.baseFilename = os.path.abspath(filename)
        self.mode = mode
        self.encoding = encoding
        self.errors = errors
        self.delay = delay
        if delay:
            Handler.__init__(self)
            self.stream = None
        else:
            StreamHandler.__init__(self, self._open())

    def close(self):
        self.acquire()
        try:
            try:
                if self.stream:
                    try:
                        self.flush()
                    finally:
                        stream = self.stream
                        self.stream = None
                        if hasattr(stream, "close"):
                            stream.close()
            finally:
                StreamHandler.close(self)
        finally:
            self.release()

    def _open(self):
        enc = self.encoding
        if enc is None and "b" not in self.mode:
            enc = "locale"
            enc = "utf-8"
        return open(self.baseFilename, self.mode, -1, enc, self.errors)

    def emit(self, record):
        if self.stream is None:
            if self.mode != "w" or not self._closed:
                self.stream = self._open()
        if self.stream:
            StreamHandler.emit(self, record)

    def __repr__(self):
        level = getLevelName(self.level)
        return "<%s %s (%s)>" % (self.__class__.__name__, self.baseFilename, level)


class _StderrHandler(StreamHandler):
    def __init__(self, level=NOTSET):
        Handler.__init__(self, level)
        self.stream = None

    def _get_stream(self):
        return sys.stderr

    def emit(self, record):
        try:
            msg = self.format(record)
            stream = sys.stderr
            stream.write(msg + self.terminator)
            stream.flush()
        except RecursionError:
            raise
        except Exception:
            self.handleError(record)


_defaultLastResort = _StderrHandler(WARNING)
lastResort = _defaultLastResort


class PlaceHolder:
    def __init__(self, alogger):
        self.loggerMap = {alogger: None}

    def append(self, alogger):
        if alogger not in self.loggerMap:
            self.loggerMap[alogger] = None


def setLoggerClass(klass):
    global _loggerClass
    if klass != Logger:
        if not issubclass(klass, Logger):
            raise TypeError("logger not derived from logging.Logger: " + klass.__name__)
    _loggerClass = klass


def getLoggerClass():
    return _loggerClass


class Manager:
    def __init__(self, rootnode):
        self.root = rootnode
        self.disable = 0
        self.emittedNoHandlerWarning = False
        self.loggerDict = {}
        self.loggerClass = None
        self.logRecordFactory = None

    @property
    def disable(self):
        return self._disable

    @disable.setter
    def disable(self, value):
        self._disable = _checkLevel(value)

    def getLogger(self, name):
        rv = None
        if not isinstance(name, str):
            raise TypeError("A logger name must be a string")
        _acquireLock()
        try:
            if name in self.loggerDict:
                rv = self.loggerDict[name]
                if isinstance(rv, PlaceHolder):
                    ph = rv
                    rv = (self.loggerClass or _loggerClass)(name)
                    rv.manager = self
                    self.loggerDict[name] = rv
                    self._fixupChildren(ph, rv)
                    self._fixupParents(rv)
            else:
                rv = (self.loggerClass or _loggerClass)(name)
                rv.manager = self
                self.loggerDict[name] = rv
                self._fixupParents(rv)
        finally:
            _releaseLock()
        return rv

    def setLoggerClass(self, klass):
        if klass != Logger:
            if not issubclass(klass, Logger):
                raise TypeError("logger not derived from logging.Logger: " + klass.__name__)
        self.loggerClass = klass

    def setLogRecordFactory(self, factory):
        self.logRecordFactory = factory

    def _fixupParents(self, alogger):
        name = alogger.name
        i = name.rfind(".")
        rv = None
        while (i > 0) and not rv:
            substr = name[:i]
            if substr not in self.loggerDict:
                self.loggerDict[substr] = PlaceHolder(alogger)
            else:
                obj = self.loggerDict[substr]
                if isinstance(obj, Logger):
                    rv = obj
                else:
                    obj.append(alogger)
            i = name.rfind(".", 0, i - 1)
        if not rv:
            rv = self.root
        alogger.parent = rv

    def _fixupChildren(self, ph, alogger):
        name = alogger.name
        namelen = len(name)
        for c in ph.loggerMap.keys():
            if c.parent.name[:namelen] != name:
                alogger.parent = c.parent
                c.parent = alogger

    def _clear_cache(self):
        _acquireLock()
        try:
            for logger in self.loggerDict.values():
                if isinstance(logger, Logger):
                    logger._cache.clear()
            self.root._cache.clear()
        finally:
            _releaseLock()


class Logger(Filterer):
    def __init__(self, name, level=NOTSET):
        Filterer.__init__(self)
        self.name = name
        self.level = _checkLevel(level)
        self.parent = None
        self.propagate = True
        self.handlers = []
        self.disabled = False
        self._cache = {}
        self.manager = _obj(None)

    def setLevel(self, level):
        self.level = _checkLevel(level)
        self.manager._clear_cache()

    def debug(self, msg, *args, **kwargs):
        if self.isEnabledFor(DEBUG):
            self._log(DEBUG, msg, args, **kwargs)

    def info(self, msg, *args, **kwargs):
        if self.isEnabledFor(INFO):
            self._log(INFO, msg, args, **kwargs)

    def warning(self, msg, *args, **kwargs):
        if self.isEnabledFor(WARNING):
            self._log(WARNING, msg, args, **kwargs)

    def warn(self, msg, *args, **kwargs):
        warnings.warn("The 'warn' method is deprecated, use 'warning' instead", DeprecationWarning, 2)
        self.warning(msg, *args, **kwargs)

    def error(self, msg, *args, **kwargs):
        if self.isEnabledFor(ERROR):
            self._log(ERROR, msg, args, **kwargs)

    def exception(self, msg, *args, exc_info=True, **kwargs):
        self.error(msg, *args, exc_info=exc_info, **kwargs)

    def critical(self, msg, *args, **kwargs):
        if self.isEnabledFor(CRITICAL):
            self._log(CRITICAL, msg, args, **kwargs)

    def fatal(self, msg, *args, **kwargs):
        self.critical(msg, *args, **kwargs)

    def log(self, level, msg, *args, **kwargs):
        if not isinstance(level, int):
            if raiseExceptions:
                raise TypeError("level must be an integer")
            else:
                return
        if self.isEnabledFor(level):
            self._log(level, msg, args, **kwargs)

    def findCaller(self, stack_info=False, stacklevel=1):
        return ("(unknown file)", 0, "(unknown function)", None)

    def makeRecord(self, name, level, fn, lno, msg, args, exc_info, func=None, extra=None, sinfo=None):
        factory = self.manager.logRecordFactory or _logRecordFactory
        rv = factory(name, level, fn, lno, msg, args, exc_info, func, sinfo)
        if extra is not None:
            for key in extra:
                if (key in ["message", "asctime"]) or (key in rv.__dict__):
                    raise KeyError("Attempt to overwrite %r in LogRecord" % key)
                setattr(rv, key, extra[key])
        return rv

    def _log(self, level, msg, args, exc_info=None, extra=None, stack_info=False, stacklevel=1):
        sinfo = None
        fn, lno, func, sinfo = self.findCaller(stack_info, stacklevel)
        if exc_info:
            if isinstance(exc_info, BaseException):
                exc_info = (type(exc_info), exc_info, None)
            elif not isinstance(exc_info, tuple):
                exc_info = sys.exc_info()
        record = self.makeRecord(self.name, level, fn, lno, msg, args, exc_info, func, extra, sinfo)
        self.handle(record)

    def handle(self, record):
        if self.disabled:
            return
        maybe_record = self.filter(record)
        if not maybe_record:
            return
        if isinstance(maybe_record, LogRecord):
            record = maybe_record
        self.callHandlers(record)

    def addHandler(self, hdlr):
        _acquireLock()
        try:
            if not (hdlr in self.handlers):
                self.handlers.append(hdlr)
        finally:
            _releaseLock()

    def removeHandler(self, hdlr):
        _acquireLock()
        try:
            if hdlr in self.handlers:
                self.handlers.remove(hdlr)
        finally:
            _releaseLock()

    def hasHandlers(self):
        c = self
        rv = False
        while c:
            if c.handlers:
                rv = True
                break
            if not c.propagate:
                break
            else:
                c = c.parent
        return rv

    def callHandlers(self, record):
        c = self
        found = 0
        while c:
            for hdlr in c.handlers:
                found = found + 1
                if record.levelno >= hdlr.level:
                    hdlr.handle(record)
            if not c.propagate:
                c = None
            else:
                c = c.parent
        if found == 0:
            if lastResort:
                if record.levelno >= lastResort.level:
                    lastResort.handle(record)
            elif raiseExceptions and not self.manager.emittedNoHandlerWarning:
                sys.stderr.write("No handlers could be found for logger \"%s\"\n" % self.name)
                self.manager.emittedNoHandlerWarning = True

    def getEffectiveLevel(self):
        logger = self
        while logger:
            if logger.level:
                return logger.level
            logger = logger.parent
        return NOTSET

    def isEnabledFor(self, level):
        if self.disabled:
            return False
        if level in self._cache:
            return self._cache[level]
        _acquireLock()
        try:
            if self.manager.disable >= level:
                is_enabled = False
                self._cache[level] = False
            else:
                is_enabled = level >= self.getEffectiveLevel()
                self._cache[level] = is_enabled
        finally:
            _releaseLock()
        return is_enabled

    def getChild(self, suffix):
        if root is not self:
            suffix = ".".join((self.name, suffix))
        return self.manager.getLogger(suffix)

    def getChildren(self):
        def _hierlevel(logger):
            if logger is root:
                return 0
            return 1 + logger.name.count(".")

        d = self.manager.loggerDict
        _acquireLock()
        try:
            return set(item for item in d.values()
                       if isinstance(item, Logger) and item.parent is self and
                       _hierlevel(item) == 1 + _hierlevel(item.parent))
        finally:
            _releaseLock()

    def __repr__(self):
        level = getLevelName(self.getEffectiveLevel())
        return "<%s %s (%s)>" % (self.__class__.__name__, self.name, level)

    def __reduce__(self):
        if getLogger(self.name) is not self:
            import pickle
            raise pickle.PicklingError("logger cannot be pickled")
        return getLogger, (self.name,)


class RootLogger(Logger):
    def __init__(self, level):
        Logger.__init__(self, "root", level)

    def __reduce__(self):
        return getLogger, ()


_loggerClass = Logger


class LoggerAdapter:
    def __init__(self, logger, extra=None, merge_extra=False):
        self.logger = logger
        self.extra = extra
        self.merge_extra = merge_extra

    def process(self, msg, kwargs):
        if self.merge_extra and "extra" in kwargs:
            merged = dict(self.extra)
            merged.update(kwargs["extra"])
            kwargs["extra"] = merged
        else:
            kwargs["extra"] = self.extra
        return msg, kwargs

    def debug(self, msg, *args, **kwargs):
        self.log(DEBUG, msg, *args, **kwargs)

    def info(self, msg, *args, **kwargs):
        self.log(INFO, msg, *args, **kwargs)

    def warning(self, msg, *args, **kwargs):
        self.log(WARNING, msg, *args, **kwargs)

    def warn(self, msg, *args, **kwargs):
        warnings.warn("The 'warn' method is deprecated, use 'warning' instead", DeprecationWarning, 2)
        self.warning(msg, *args, **kwargs)

    def error(self, msg, *args, **kwargs):
        self.log(ERROR, msg, *args, **kwargs)

    def exception(self, msg, *args, exc_info=True, **kwargs):
        self.log(ERROR, msg, *args, exc_info=exc_info, **kwargs)

    def critical(self, msg, *args, **kwargs):
        self.log(CRITICAL, msg, *args, **kwargs)

    def log(self, level, msg, *args, **kwargs):
        if self.isEnabledFor(level):
            msg, kwargs = self.process(msg, kwargs)
            self.logger.log(level, msg, *args, **kwargs)

    def isEnabledFor(self, level):
        return self.logger.isEnabledFor(level)

    def setLevel(self, level):
        self.logger.setLevel(level)

    def getEffectiveLevel(self):
        return self.logger.getEffectiveLevel()

    def hasHandlers(self):
        return self.logger.hasHandlers()

    @property
    def manager(self):
        return self.logger.manager

    @property
    def name(self):
        return self.logger.name

    def __repr__(self):
        logger = self.logger
        level = getLevelName(logger.getEffectiveLevel())
        return "<%s %s (%s)>" % (self.__class__.__name__, logger.name, level)


root = RootLogger(WARNING)
_manager = Manager(root)
root.manager = _manager


def basicConfig(**kwargs):
    _acquireLock()
    try:
        force = kwargs.pop("force", False)
        encoding = kwargs.pop("encoding", None)
        errors = kwargs.pop("errors", "backslashreplace")
        if force:
            for h in root.handlers[:]:
                root.removeHandler(h)
                h.close()
        if len(root.handlers) == 0:
            handlers = kwargs.pop("handlers", None)
            if handlers is None:
                if "stream" in kwargs and "filename" in kwargs:
                    raise ValueError("'stream' and 'filename' should not be specified together")
            else:
                if "stream" in kwargs or "filename" in kwargs:
                    raise ValueError("'stream' or 'filename' should not be specified together with 'handlers'")
            if handlers is None:
                filename = kwargs.pop("filename", None)
                mode = kwargs.pop("filemode", "a")
                if filename:
                    if "b" in mode:
                        errors = None
                    else:
                        encoding = encoding if encoding is not None else "utf-8"
                    h = FileHandler(filename, mode, encoding=encoding, errors=errors)
                else:
                    stream = kwargs.pop("stream", None)
                    h = StreamHandler(stream)
                handlers = [h]
            dfs = kwargs.pop("datefmt", None)
            style = kwargs.pop("style", "%")
            if style not in _STYLES:
                raise ValueError("Style must be one of: %s" % ",".join(_STYLES.keys()))
            fs = kwargs.pop("format", _STYLES[style][1])
            fmt = Formatter(fs, dfs, style)
            for h in handlers:
                if h.formatter is None:
                    h.setFormatter(fmt)
                root.addHandler(h)
            level = kwargs.pop("level", None)
            if level is not None:
                root.setLevel(level)
            if kwargs:
                keys = ", ".join(kwargs.keys())
                raise ValueError("Unrecognised argument(s): %s" % keys)
    finally:
        _releaseLock()


def getLogger(name=None):
    if not name or isinstance(name, str) and name == root.name:
        return root
    return _manager.getLogger(name)


def critical(msg, *args, **kwargs):
    if len(root.handlers) == 0:
        basicConfig()
    root.critical(msg, *args, **kwargs)


def fatal(msg, *args, **kwargs):
    critical(msg, *args, **kwargs)


def error(msg, *args, **kwargs):
    if len(root.handlers) == 0:
        basicConfig()
    root.error(msg, *args, **kwargs)


def exception(msg, *args, exc_info=True, **kwargs):
    error(msg, *args, exc_info=exc_info, **kwargs)


def warning(msg, *args, **kwargs):
    if len(root.handlers) == 0:
        basicConfig()
    root.warning(msg, *args, **kwargs)


def warn(msg, *args, **kwargs):
    warnings.warn("The 'warn' function is deprecated, use 'warning' instead", DeprecationWarning, 2)
    warning(msg, *args, **kwargs)


def info(msg, *args, **kwargs):
    if len(root.handlers) == 0:
        basicConfig()
    root.info(msg, *args, **kwargs)


def debug(msg, *args, **kwargs):
    if len(root.handlers) == 0:
        basicConfig()
    root.debug(msg, *args, **kwargs)


def log(level, msg, *args, **kwargs):
    if len(root.handlers) == 0:
        basicConfig()
    root.log(level, msg, *args, **kwargs)


def disable(level=CRITICAL):
    root.manager.disable = level
    root.manager._clear_cache()


def shutdown(handlerList=_handlerList):
    for h in handlerList[:]:
        try:
            h.acquire()
            try:
                h.flush()
                h.close()
            finally:
                h.release()
        except (OSError, ValueError):
            pass
        except Exception:
            if raiseExceptions:
                raise


import atexit
atexit.register(shutdown)


class NullHandler(Handler):
    def handle(self, record):
        pass

    def emit(self, record):
        pass

    def createLock(self):
        self.lock = None

    def _at_fork_reinit(self):
        pass


_warnings_showwarning = None


def captureWarnings(capture):
    pass
