"""shlex: a simple lexical analyzer for Unix shell-like syntaxes."""

import re
import os
from collections import deque

_find_unsafe = re.compile(r"[^\w@%+=:,./-]", re.ASCII).search


def quote(s):
    if not s:
        return "''"
    if _find_unsafe(s) is None:
        return s
    return "'" + s.replace("'", "'\"'\"'") + "'"


def join(split_command):
    return " ".join(quote(arg) for arg in split_command)


class shlex:
    def __init__(self, instream=None, infile=None, posix=False, punctuation_chars=False):
        if isinstance(instream, str):
            self._text = instream
        elif instream is None:
            self._text = ""
        else:
            self._text = instream.read()
        self._pos = 0
        self.posix = posix
        self.eof = None
        if posix:
            self.eof = None
        else:
            self.eof = ""
        self.commenters = "#"
        self.wordchars = "abcdfeghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
        if self.posix:
            self.wordchars += "ßàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþÿÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ"
        self.whitespace = " \t\r\n"
        self.whitespace_split = False
        self.quotes = "'\""
        self.escape = "\\"
        self.escapedquotes = "\""
        self.state = " "
        self.pushback = deque()
        self.lineno = 1
        self.debug = 0
        self.token = ""
        self.filestack = deque()
        self.source = None
        self.infile = infile
        self.instream = instream
        if not punctuation_chars:
            punctuation_chars = ""
        elif punctuation_chars is True:
            punctuation_chars = "();<>|&"
        self._punctuation_chars = punctuation_chars
        if punctuation_chars:
            self._pushback_chars = deque()
            self.wordchars += "~-./*?="
            remove = "".join(c for c in self._punctuation_chars if c not in "~-./*?=")
            self.wordchars = "".join(c for c in self.wordchars if c not in remove)

    @property
    def punctuation_chars(self):
        return self._punctuation_chars

    def push_token(self, tok):
        self.pushback.appendleft(tok)

    def pop_token(self):
        return self.pushback.popleft()

    def get_token(self):
        if self.pushback:
            return self.pushback.popleft()
        raw = self.read_token()
        return raw

    def _getc(self):
        if self._punctuation_chars and self._pushback_chars:
            return self._pushback_chars.pop()
        if self._pos >= len(self._text):
            return ""
        c = self._text[self._pos]
        self._pos += 1
        return c

    def read_token(self):
        quoted = False
        escapedstate = " "
        while True:
            if self._punctuation_chars and self._pushback_chars:
                nextchar = self._pushback_chars.pop()
            else:
                nextchar = self._getc()
            if nextchar == "\n":
                self.lineno += 1
            if self.state is None:
                self.token = ""
                break
            elif self.state == " ":
                if not nextchar:
                    self.state = None
                    break
                elif nextchar in self.whitespace:
                    if self.token or (self.posix and quoted):
                        break
                    else:
                        continue
                elif nextchar in self.commenters:
                    idx = self._text.find("\n", self._pos)
                    if idx < 0:
                        self._pos = len(self._text)
                    else:
                        self._pos = idx + 1
                    self.lineno += 1
                elif self.posix and nextchar in self.escape:
                    escapedstate = "a"
                    self.state = nextchar
                elif nextchar in self.wordchars:
                    self.token = nextchar
                    self.state = "a"
                elif nextchar in self._punctuation_chars:
                    self.token = nextchar
                    self.state = "c"
                elif nextchar in self.quotes:
                    if not self.posix:
                        self.token = nextchar
                    self.state = nextchar
                elif self.whitespace_split:
                    self.token = nextchar
                    self.state = "a"
                else:
                    self.token = nextchar
                    if self.token or (self.posix and quoted):
                        break
                    else:
                        continue
            elif self.state in self.quotes:
                quoted = True
                if not nextchar:
                    raise ValueError("No closing quotation")
                if nextchar == self.state:
                    if not self.posix:
                        self.token += nextchar
                        self.state = " "
                        break
                    else:
                        self.state = "a"
                elif self.posix and nextchar in self.escape and self.state in self.escapedquotes:
                    escapedstate = self.state
                    self.state = nextchar
                else:
                    self.token += nextchar
            elif self.state in self.escape:
                if not nextchar:
                    raise ValueError("No escaped character")
                if escapedstate in self.quotes and nextchar != self.state and nextchar != escapedstate:
                    self.token += self.state
                self.token += nextchar
                self.state = escapedstate
            elif self.state in ("a", "c"):
                if not nextchar:
                    self.state = None
                    break
                elif nextchar in self.whitespace:
                    self.state = " "
                    if self.token or (self.posix and quoted):
                        break
                    else:
                        continue
                elif nextchar in self.commenters:
                    idx = self._text.find("\n", self._pos)
                    if idx < 0:
                        self._pos = len(self._text)
                    else:
                        self._pos = idx + 1
                    self.lineno += 1
                    if self.posix:
                        self.state = " "
                        if self.token or (self.posix and quoted):
                            break
                        else:
                            continue
                elif self.state == "c":
                    if nextchar in self._punctuation_chars:
                        self.token += nextchar
                    else:
                        if nextchar not in self.whitespace:
                            self._pushback_chars.append(nextchar)
                        self.state = " "
                        break
                elif self.posix and nextchar in self.quotes:
                    self.state = nextchar
                elif self.posix and nextchar in self.escape:
                    escapedstate = "a"
                    self.state = nextchar
                elif nextchar in self.wordchars or nextchar in self.quotes or (self.whitespace_split and nextchar not in self._punctuation_chars):
                    self.token += nextchar
                else:
                    if self._punctuation_chars:
                        self._pushback_chars.append(nextchar)
                    else:
                        self._pos -= 1
                    self.state = " "
                    if self.token or (self.posix and quoted):
                        break
                    else:
                        continue
        result = self.token
        self.token = ""
        if self.posix and not quoted and result == "":
            result = None
        return result

    def __iter__(self):
        return self

    def __next__(self):
        token = self.get_token()
        if token == self.eof:
            raise StopIteration
        return token


def split(s, comments=False, posix=True):
    if s is None:
        raise ValueError("s argument must not be None")
    lex = shlex(s, posix=posix)
    lex.whitespace_split = True
    if not comments:
        lex.commenters = ""
    out = []
    while True:
        tok = lex.get_token()
        if tok == lex.eof:
            break
        out.append(tok)
    return out
