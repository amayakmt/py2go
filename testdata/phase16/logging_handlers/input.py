import logging
import logging.handlers
import os
import queue
import tempfile

d = tempfile.mkdtemp()
path = os.path.join(d, "app.log")
lg = logging.getLogger("rot")
lg.setLevel(logging.DEBUG)
lg.propagate = False
h = logging.handlers.RotatingFileHandler(path, maxBytes=60, backupCount=2)
h.setFormatter(logging.Formatter("%(levelname)s %(message)s"))
lg.addHandler(h)
for i in range(8):
    lg.info("message number %d", i)
h.close()
for name in sorted(os.listdir(d)):
    with open(os.path.join(d, name)) as f:
        print(name, repr(f.read()))

mem = logging.handlers.MemoryHandler(capacity=3, flushLevel=logging.ERROR, target=None)
class Sink(logging.Handler):
    def __init__(self):
        logging.Handler.__init__(self)
        self.out = []
    def emit(self, record):
        self.out.append(self.format(record))
sink = Sink()
sink.setFormatter(logging.Formatter("%(message)s"))
mem.setTarget(sink)
ml = logging.getLogger("mem")
ml.propagate = False
ml.setLevel(logging.DEBUG)
ml.addHandler(mem)
ml.info("a")
ml.info("b")
print(sink.out, len(mem.buffer))
ml.info("c")
print(sink.out, len(mem.buffer))
ml.info("d")
ml.error("boom")
print(sink.out, len(mem.buffer))

q = queue.Queue()
qh = logging.handlers.QueueHandler(q)
ql = logging.getLogger("queued")
ql.propagate = False
ql.setLevel(logging.INFO)
ql.addHandler(qh)
ql.info("to queue %s", "x")
rec = q.get_nowait()
print(rec.getMessage(), rec.levelname, rec.name, rec.args, q.empty())
sink2 = Sink()
sink2.setFormatter(logging.Formatter("%(levelname)s:%(message)s"))
listener = logging.handlers.QueueListener(q, sink2)
listener.start()
ql.warning("listened")
ql.error("again")
listener.stop()
print(sink2.out)

wf = logging.handlers.WatchedFileHandler(os.path.join(d, "watched.log"))
wl = logging.getLogger("watched")
wl.propagate = False
wl.addHandler(wf)
wl.warning("one")
os.remove(os.path.join(d, "watched.log"))
wl.warning("two")
wf.close()
print(open(os.path.join(d, "watched.log")).read(), end="")

th = logging.handlers.TimedRotatingFileHandler(os.path.join(d, "timed.log"), when="midnight", backupCount=1)
print(th.when, th.interval, th.suffix, th.backupCount, th.rolloverAt > 0)
th.close()
try:
    logging.handlers.TimedRotatingFileHandler(os.path.join(d, "x.log"), when="X")
except ValueError as e:
    print(e)
import shutil
shutil.rmtree(d)
