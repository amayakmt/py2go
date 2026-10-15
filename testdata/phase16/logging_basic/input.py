import logging
import sys
import io

print(logging.getLevelName(10), logging.getLevelName("WARNING"), logging.getLevelName(15), logging.WARN, logging.FATAL)
stream = io.StringIO()
h = logging.StreamHandler(stream)
h.setFormatter(logging.Formatter("%(levelname)s:%(name)s:%(message)s"))
log = logging.getLogger("app.sub")
log.addHandler(h)
log.setLevel(logging.DEBUG)
log.debug("debug %s", "msg")
log.info("info %d + %d", 1, 2)
log.warning("warn")
log.error("error %(a)s", {"a": "dict-args"})
log.critical("crit")
log.log(25, "custom level")
print(stream.getvalue(), end="")
print(log, repr(h), log.name, log.parent, log.getEffectiveLevel(), log.isEnabledFor(5), logging.getLogger("app.sub") is log)
print(logging.getLogger("app"), logging.getLogger() is logging.root, logging.getLogger("").name, log.parent.name, logging.getLogger("app").getChild("x"))

h.setLevel(logging.ERROR)
stream.truncate(0)
stream.seek(0)
log.info("hidden")
log.error("shown")
print(repr(stream.getvalue()))
log.removeHandler(h)

class Capture(logging.Handler):
    def __init__(self):
        logging.Handler.__init__(self)
        self.records = []

    def emit(self, record):
        self.records.append(record)

cap = Capture()
parent = logging.getLogger("parent")
parent.addHandler(cap)
parent.setLevel(logging.INFO)
child = logging.getLogger("parent.child")
child.info("from child %s", 42)
child.debug("not recorded")
child.propagate = False
child.info("lost")
rec = cap.records[0]
print(len(cap.records), rec.name, rec.levelname, rec.levelno, rec.getMessage(), rec.msg, rec.args)

class Flt(logging.Filter):
    def filter(self, record):
        return "secret" not in record.getMessage()
cap2 = Capture()
cap2.addFilter(Flt())
lg = logging.getLogger("filtered")
lg.addHandler(cap2)
lg.setLevel(logging.DEBUG)
lg.info("public")
lg.info("a secret")
print([r.getMessage() for r in cap2.records])

fmt = logging.Formatter("{levelname} - {message}", style="{")
r = logging.makeLogRecord({"msg": "hi %s", "args": ("x",), "levelname": "INFO", "levelno": 20, "name": "n"})
print(fmt.format(r), logging.Formatter("%(message)s|%(name)s").format(r), logging.Formatter("$levelname $message", style="$").format(r))
try:
    logging.Formatter("%(nope)s").format(r)
except ValueError as e:
    print(e)
try:
    logging.Formatter("no fields", validate=True)
except ValueError as e:
    print(e)
try:
    logging.getLogger("x").setLevel("BOGUS")
except ValueError as e:
    print(e)
try:
    logging.getLogger("x").setLevel(1.5)
except TypeError as e:
    print(e)

adapter = logging.LoggerAdapter(lg, {"user": "bob"})
cap3 = Capture()
lg.addHandler(cap3)
adapter.warning("adapted %s", 1)
print(cap3.records[0].user, cap3.records[0].getMessage(), adapter)

logging.disable(logging.CRITICAL)
lg.critical("silenced")
print(len(cap3.records))
logging.disable(logging.NOTSET)

logging.basicConfig(stream=sys.stdout, level=logging.INFO, format="%(levelname)s %(message)s")
logging.info("root info")
logging.debug("root debug")
logging.warning("root %s", "warn")
print(logging.getLogger().level, logging.getLogger().handlers[0].formatter._fmt)
e = Exception("oops")
logging.error("with exc", exc_info=e)

nh = logging.NullHandler()
logging.getLogger("nowhere").addHandler(nh)
logging.getLogger("nowhere").error("goes nowhere")
logging.getLogger("lastresort").propagate = False
logging.getLogger("lastresort").warning("to stderr via lastResort")
logging.shutdown()
