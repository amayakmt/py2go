"""fnmatch: Unix shell style filename pattern matching."""

import os
import re

_cache = {}


def translate(pat):
    STAR = object()
    parts = []
    i = 0
    n = len(pat)
    res = []
    while i < n:
        c = pat[i]
        i = i + 1
        if c == "*":
            if not res or res[-1] != "*":
                res.append("*")
        elif c == "?":
            res.append(".")
        elif c == "[":
            j = i
            if j < n and pat[j] == "!":
                j = j + 1
            if j < n and pat[j] == "]":
                j = j + 1
            while j < n and pat[j] != "]":
                j = j + 1
            if j >= n:
                res.append("\\[")
            else:
                stuff = pat[i:j]
                if "-" not in stuff:
                    stuff = stuff.replace("\\", "\\\\")
                else:
                    chunks = []
                    k = i + 2 if pat[i] == "!" else i + 1
                    while True:
                        k = pat.find("-", k, j)
                        if k < 0:
                            break
                        chunks.append(pat[i:k])
                        i = k + 1
                        k = k + 3
                    chunk = pat[i:j]
                    if chunk:
                        chunks.append(chunk)
                    else:
                        chunks[-1] += "-"
                    k = len(chunks) - 1
                    while k > 0:
                        if chunks[k - 1][-1] > chunks[k][0]:
                            chunks[k - 1] = chunks[k - 1][:-1] + chunks[k][1:]
                            del chunks[k]
                        k -= 1
                    stuff = "-".join(s.replace("\\", "\\\\").replace("-", "\\-") for s in chunks)
                stuff = re.sub(r"([&~|])", r"\\\1", stuff)
                i = j + 1
                if not stuff:
                    res.append("(?!)")
                elif stuff == "!":
                    res.append(".")
                else:
                    if stuff[0] == "!":
                        stuff = "^" + stuff[1:]
                    elif stuff[0] in ("^", "["):
                        stuff = "\\" + stuff
                    res.append("[" + stuff + "]")
        else:
            res.append(re.escape(c))
    out = ""
    for piece in res:
        if piece == "*":
            out += ".*"
        else:
            out += piece
    return "(?s:" + out + ")\\Z"


def _compile(pat):
    r = _cache.get(pat)
    if r is None:
        r = re.compile(translate(pat))
        _cache[pat] = r
    return r


def filter(names, pat):
    result = []
    pat = os.path.normcase(pat)
    match = _compile(pat).match
    for name in names:
        if match(os.path.normcase(name)):
            result.append(name)
    return result


def fnmatchcase(name, pat):
    return _compile(pat).match(name) is not None


def fnmatch(name, pat):
    return fnmatchcase(os.path.normcase(name), os.path.normcase(pat))
