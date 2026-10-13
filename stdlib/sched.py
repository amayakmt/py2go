"""sched: a generally useful event scheduler class."""

import time
import heapq
import threading
from time import monotonic as _time

_sentinel = object()


class Event:
    _fields = ("time", "priority", "sequence", "action", "argument", "kwargs")

    def __init__(self, time, priority, sequence, action, argument, kwargs):
        self.time = time
        self.priority = priority
        self.sequence = sequence
        self.action = action
        self.argument = argument
        self.kwargs = kwargs

    def _key(self):
        return (self.time, self.priority, self.sequence)

    def __eq__(self, other):
        return self._key() == other._key()

    def __lt__(self, other):
        return self._key() < other._key()

    def __le__(self, other):
        return self._key() <= other._key()

    def __gt__(self, other):
        return self._key() > other._key()

    def __ge__(self, other):
        return self._key() >= other._key()

    def __iter__(self):
        return iter((self.time, self.priority, self.sequence, self.action, self.argument, self.kwargs))

    def __getitem__(self, i):
        return (self.time, self.priority, self.sequence, self.action, self.argument, self.kwargs)[i]

    def __len__(self):
        return 6

    def __repr__(self):
        return "Event(time=" + repr(self.time) + ", priority=" + repr(self.priority) + ", sequence=" + repr(self.sequence) + ", action=" + repr(self.action) + ", argument=" + repr(self.argument) + ", kwargs=" + repr(self.kwargs) + ")"


class scheduler:
    def __init__(self, timefunc=_time, delayfunc=time.sleep):
        self._queue = []
        self._lock = threading.RLock()
        self.timefunc = timefunc
        self.delayfunc = delayfunc
        self._sequence_generator = 0

    def enterabs(self, time, priority, action, argument=(), kwargs=_sentinel):
        if kwargs is _sentinel:
            kwargs = {}
        with self._lock:
            self._sequence_generator += 1
            event = Event(time, priority, self._sequence_generator, action, argument, kwargs)
            heapq.heappush(self._queue, event)
        return event

    def enter(self, delay, priority, action, argument=(), kwargs=_sentinel):
        time = self.timefunc() + delay
        return self.enterabs(time, priority, action, argument, kwargs)

    def cancel(self, event):
        with self._lock:
            self._queue.remove(event)
            heapq.heapify(self._queue)

    def empty(self):
        with self._lock:
            return not self._queue

    def run(self, blocking=True):
        lock = self._lock
        q = self._queue
        delayfunc = self.delayfunc
        timefunc = self.timefunc
        pop = heapq.heappop
        while True:
            with lock:
                if not q:
                    break
                event = q[0]
                now = timefunc()
                if event.time > now:
                    delay = True
                else:
                    delay = False
                    pop(q)
            if delay:
                if not blocking:
                    return event.time - now
                delayfunc(event.time - now)
            else:
                event.action(*event.argument, **event.kwargs)
                delayfunc(0)

    @property
    def queue(self):
        with self._lock:
            events = self._queue[:]
        out = []
        while events:
            out.append(heapq.heappop(events))
        return out
