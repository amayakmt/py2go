"""gc: the garbage collector interface (the Go collector runs on its own)."""

DEBUG_STATS = 1
DEBUG_COLLECTABLE = 2
DEBUG_UNCOLLECTABLE = 4
DEBUG_SAVEALL = 32
DEBUG_LEAK = 38

garbage = []
callbacks = []
_enabled = True
_debug = 0
_threshold = (700, 10, 10)


def enable():
    global _enabled
    _enabled = True


def disable():
    global _enabled
    _enabled = False


def isenabled():
    return _enabled


def collect(generation=2):
    if generation not in (0, 1, 2):
        raise ValueError("invalid generation")
    return 0


def set_debug(flags):
    global _debug
    _debug = flags


def get_debug():
    return _debug


def get_threshold():
    return _threshold


def set_threshold(threshold0, threshold1=10, threshold2=10):
    global _threshold
    _threshold = (threshold0, threshold1, threshold2)


def get_count():
    return (0, 0, 0)


def get_stats():
    return [{"collections": 0, "collected": 0, "uncollectable": 0}, {"collections": 0, "collected": 0, "uncollectable": 0},
            {"collections": 0, "collected": 0, "uncollectable": 0}]


def get_objects(generation=None):
    return []


def get_referrers(*objs):
    return []


def get_referents(*objs):
    return []


def is_tracked(obj):
    return not (obj is None or isinstance(obj, (int, str, float, bool, bytes)))


def is_finalized(obj):
    return False


def freeze():
    pass


def unfreeze():
    pass


def get_freeze_count():
    return 0
