"""dbm: a generic interface to database-like modules (only dbm.dumb is available)."""
from . import dumb

__all__ = ["open", "whichdb", "error"]


class error(Exception):
    pass


def open(file, flag="r", mode=0o666):
    if flag not in ("r", "w", "c", "n"):
        raise ValueError("Flag must be one of 'r', 'w', 'c', or 'n'")
    import os
    path = os.fspath(file)
    if flag in ("r", "w"):
        if not os.path.exists(path + ".dat") and not os.path.exists(path + ".dir"):
            raise error("db file doesn't exist; use 'c' or 'n' flag to create a new db")
    if flag == "r":
        return dumb._Database(path, mode, "r")
    return dumb._Database(path, mode, "w" if flag == "w" else flag)


def whichdb(filename):
    import os
    if os.path.exists(filename + ".dat") and os.path.exists(filename + ".dir"):
        return "dbm.dumb"
    return None
