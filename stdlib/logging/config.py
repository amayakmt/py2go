"""logging.config: configuration of the logging module from dictionaries."""
import logging
import logging.handlers
import sys

DEFAULT_LOGGING_CONFIG_PORT = 9030
RESET_ERROR = 54


def _resolve_class(name):
    table = {
        "logging.StreamHandler": logging.StreamHandler,
        "logging.FileHandler": logging.FileHandler,
        "logging.NullHandler": logging.NullHandler,
        "logging.handlers.RotatingFileHandler": logging.handlers.RotatingFileHandler,
        "logging.handlers.TimedRotatingFileHandler": logging.handlers.TimedRotatingFileHandler,
        "logging.handlers.WatchedFileHandler": logging.handlers.WatchedFileHandler,
        "logging.handlers.MemoryHandler": logging.handlers.MemoryHandler,
        "logging.handlers.QueueHandler": logging.handlers.QueueHandler,
        "logging.handlers.BufferingHandler": logging.handlers.BufferingHandler,
        "logging.Formatter": logging.Formatter,
        "logging.Filter": logging.Filter,
    }
    if name in table:
        return table[name]
    raise ValueError("Unable to resolve class %r" % name)


def _ext(value):
    if isinstance(value, str) and value.startswith("ext://"):
        target = value[6:]
        if target == "sys.stdout":
            return sys.stdout
        if target == "sys.stderr":
            return sys.stderr
        if target == "sys.stdin":
            return sys.stdin
        raise ValueError("Cannot resolve %r" % value)
    return value


def _level(value):
    return logging._checkLevel(value)


def _make_formatter(cfg):
    if "()" in cfg:
        factory = cfg["()"]
        if isinstance(factory, str):
            factory = _resolve_class(factory)
        kwargs = {}
        for k in cfg:
            if k != "()":
                kwargs[k] = cfg[k]
        return factory(**kwargs)
    fmt = cfg.get("format", None)
    dfmt = cfg.get("datefmt", None)
    style = cfg.get("style", "%")
    validate = cfg.get("validate", True)
    defaults = cfg.get("defaults", None)
    cls = logging.Formatter
    if "class" in cfg:
        cls = _resolve_class(cfg["class"])
    return cls(fmt, dfmt, style, validate, defaults=defaults)


def _make_filter(cfg):
    if "()" in cfg:
        factory = cfg["()"]
        if isinstance(factory, str):
            factory = _resolve_class(factory)
        kwargs = {}
        for k in cfg:
            if k != "()":
                kwargs[k] = cfg[k]
        return factory(**kwargs)
    return logging.Filter(cfg.get("name", ""))


def _make_handler(cfg, formatters, filters, handlers):
    cfg = dict(cfg)
    fmt = cfg.pop("formatter", None)
    level = cfg.pop("level", None)
    flts = cfg.pop("filters", None)
    factory = None
    if "()" in cfg:
        factory = cfg.pop("()")
        if isinstance(factory, str):
            factory = _resolve_class(factory)
    else:
        cname = cfg.pop("class")
        factory = _resolve_class(cname)
    target = cfg.pop("target", None)
    kwargs = {}
    for k in cfg:
        kwargs[k] = _ext(cfg[k])
    if target is not None and isinstance(target, str):
        if target not in handlers:
            raise ValueError("Unable to set target handler %r" % target)
        kwargs["target"] = handlers[target]
    try:
        h = factory(**kwargs)
    except TypeError as te:
        if "'stream'" in str(te) and "stream" in kwargs:
            kwargs["strm"] = kwargs.pop("stream")
            h = factory(**kwargs)
        else:
            raise
    if fmt is not None:
        if fmt not in formatters:
            raise ValueError("Unable to set formatter %r" % fmt)
        h.setFormatter(formatters[fmt])
    if level is not None:
        h.setLevel(_level(level))
    if flts:
        for f in flts:
            h.addFilter(filters[f] if isinstance(f, str) else f)
    return h


def _configure_logger(logger, cfg, handlers, filters, incremental):
    level = cfg.get("level", None)
    if level is not None:
        logger.setLevel(_level(level))
    if not incremental:
        for h in logger.handlers[:]:
            logger.removeHandler(h)
        for name in cfg.get("handlers", []):
            if name not in handlers:
                raise ValueError("Unable to add handler %r" % name)
            logger.addHandler(handlers[name])
        for f in cfg.get("filters", []):
            logger.addFilter(filters[f] if isinstance(f, str) else f)
    if "propagate" in cfg:
        logger.propagate = cfg["propagate"]


def dictConfig(config):
    version = config.get("version", None)
    if version is None:
        raise ValueError("dictionary doesn't specify a version")
    if version != 1:
        raise ValueError("Unsupported version: %s" % version)
    incremental = config.get("incremental", False)
    disable_existing = config.get("disable_existing_loggers", True)
    logging._acquireLock()
    try:
        formatters = {}
        filters = {}
        handlers = {}
        if incremental:
            for name in config.get("handlers", {}):
                h = logging._handlers.get(name)
                if h is None:
                    raise ValueError("No handler found with name %r" % name)
                hc = config["handlers"][name]
                if "level" in hc:
                    h.setLevel(_level(hc["level"]))
            for name in config.get("loggers", {}):
                _configure_logger(logging.getLogger(name), config["loggers"][name], handlers, filters, True)
            if "root" in config:
                _configure_logger(logging.root, config["root"], handlers, filters, True)
            return
        logging._handlers.clear()
        del logging._handlerList[:]
        for name in config.get("formatters", {}):
            try:
                formatters[name] = _make_formatter(config["formatters"][name])
            except Exception as e:
                raise ValueError("Unable to configure formatter %r" % name) from e
        for name in config.get("filters", {}):
            try:
                filters[name] = _make_filter(config["filters"][name])
            except Exception as e:
                raise ValueError("Unable to configure filter %r" % name) from e
        pending = dict(config.get("handlers", {}))
        # handlers that name a target are made after their targets
        made = True
        while pending and made:
            made = False
            for name in list(pending):
                tgt = pending[name].get("target", None)
                if isinstance(tgt, str) and tgt in pending and tgt not in handlers:
                    continue
                try:
                    h = _make_handler(pending[name], formatters, filters, handlers)
                except Exception as e:
                    raise ValueError("Unable to configure handler %r" % name) from e
                h.name = name
                handlers[name] = h
                del pending[name]
                made = True
        if pending:
            raise ValueError("Unable to configure handler %r" % list(pending)[0])
        root = logging.root
        existing = sorted(root.manager.loggerDict.keys())
        child_loggers = []
        loggers = config.get("loggers", {})
        for name in loggers:
            if name in existing:
                i = existing.index(name) + 1
                prefixed = name + "."
                pflen = len(prefixed)
                num_existing = len(existing)
                while i < num_existing:
                    if existing[i][:pflen] == prefixed:
                        child_loggers.append(existing[i])
                    i += 1
                existing.remove(name)
            try:
                _configure_logger(logging.getLogger(name), loggers[name], handlers, filters, False)
            except Exception as e:
                raise ValueError("Unable to configure logger %r" % name) from e
        for log in existing:
            logger = root.manager.loggerDict[log]
            if isinstance(logger, logging.Logger):
                if log in child_loggers:
                    logger.level = logging.NOTSET
                    logger.handlers = []
                    logger.propagate = True
                else:
                    logger.disabled = disable_existing
            else:
                pass
        if "root" in config:
            try:
                _configure_logger(root, config["root"], handlers, filters, False)
            except Exception as e:
                raise ValueError("Unable to configure root logger") from e
    finally:
        logging._releaseLock()


def fileConfig(fname, defaults=None, disable_existing_loggers=True, encoding=None):
    raise NotImplementedError("fileConfig needs configparser-based handler lookup by name")
