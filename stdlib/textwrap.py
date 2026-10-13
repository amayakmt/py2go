"""textwrap: wrapping and filling text."""


def _wrap_chunks(words, width, initial_indent, subsequent_indent, break_long_words, break_on_hyphens):
    lines = []
    cur = []
    cur_len = 0
    indent = initial_indent
    i = 0
    chunks = list(words)
    chunks.reverse()
    while chunks:
        cur = []
        cur_len = 0
        indent = subsequent_indent if lines else initial_indent
        avail = width - len(indent)
        if chunks[-1].strip() == "" and lines:
            del chunks[-1]
        while chunks:
            l = len(chunks[-1])
            if cur_len + l <= avail:
                cur.append(chunks.pop())
                cur_len += l
            else:
                break
        if chunks and len(chunks[-1]) > avail:
            if break_long_words:
                space_left = avail - cur_len
                if avail < 1:
                    space_left = 1
                chunk = chunks[-1]
                cur.append(chunk[:space_left])
                chunks[-1] = chunk[space_left:]
            elif not cur:
                cur.append(chunks.pop())
        if cur and cur[-1].strip() == "":
            cur_len -= len(cur[-1])
            del cur[-1]
        if cur:
            lines.append(indent + "".join(cur))
    return lines


def _split(text, break_on_hyphens):
    chunks = []
    word = ""
    space = ""
    for ch in text:
        if ch in " \t\n\r\x0b\x0c":
            if word:
                chunks.append(word)
                word = ""
            space += ch
        else:
            if space:
                chunks.append(space)
                space = ""
            word += ch
            if break_on_hyphens and ch == "-" and len(word) > 1 and word[-2].isalnum():
                chunks.append(word)
                word = ""
    if word:
        chunks.append(word)
    if space:
        chunks.append(space)
    return chunks


def wrap(text, width=70, *, initial_indent="", subsequent_indent="", expand_tabs=True, replace_whitespace=True, fix_sentence_endings=False, break_long_words=True, drop_whitespace=True, break_on_hyphens=True, tabsize=8, max_lines=None, placeholder=" [...]"):
    if expand_tabs:
        text = text.expandtabs(tabsize)
    if replace_whitespace:
        text = text.translate({9: 32, 10: 32, 11: 32, 12: 32, 13: 32})
    chunks = _split(text, break_on_hyphens)
    lines = _wrap_chunks(chunks, width, initial_indent, subsequent_indent, break_long_words, break_on_hyphens)
    if max_lines is not None and len(lines) > max_lines:
        lines = lines[:max_lines]
        last = lines[-1].rstrip()
        lines[-1] = last + placeholder
    return lines


def fill(text, width=70, **kwargs):
    return "\n".join(wrap(text, width, **kwargs))


def shorten(text, width, *, placeholder=" [...]", **kwargs):
    text = " ".join(text.split())
    if len(text) <= width:
        return text
    words = text.split(" ")
    out = ""
    for w in words:
        candidate = w if not out else out + " " + w
        if len(candidate) + len(placeholder) > width:
            break
        out = candidate
    if not out:
        return placeholder.lstrip()
    return out + placeholder


def dedent(text):
    lines = text.split("\n")
    margin = None
    for line in lines:
        stripped = line.lstrip(" \t")
        if not stripped:
            continue
        indent = line[:len(line) - len(stripped)]
        if margin is None:
            margin = indent
        elif indent.startswith(margin):
            pass
        elif margin.startswith(indent):
            margin = indent
        else:
            common = ""
            for a, b in zip(margin, indent):
                if a != b:
                    break
                common += a
            margin = common
    out = []
    for line in lines:
        if not line.strip(" \t"):
            out.append("")
        elif margin:
            out.append(line[len(margin):])
        else:
            out.append(line)
    return "\n".join(out)


def indent(text, prefix, predicate=None):
    out = []
    for line in text.splitlines(True):
        if (predicate(line) if predicate is not None else line.strip()):
            out.append(prefix + line)
        else:
            out.append(line)
    return "".join(out)
