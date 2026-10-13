"""configparser: configuration file parser (INI style)."""

import re
import sys
from collections import ChainMap

DEFAULTSECT = "DEFAULT"
MAX_INTERPOLATION_DEPTH = 10

_UNSET = object()


def _obj(x):
    return x


_SECT_TMPL = r"\[(?P<header>.+)\]"
_OPT_TMPL = r"(?P<option>.*?)\s*(?P<vi>{delim})\s*(?P<value>.*)$"
_OPT_NV_TMPL = r"(?P<option>.*?)\s*(?:(?P<vi>{delim})\s*(?P<value>.*))?$"


class Error(Exception):
    def __init__(self, msg=""):
        self.message = msg
        Exception.__init__(self, msg)

    def __repr__(self):
        return self.message

    def __str__(self):
        return self.message


class NoSectionError(Error):
    def __init__(self, section):
        Error.__init__(self, "No section: " + repr(section))
        self.section = section


class DuplicateSectionError(Error):
    def __init__(self, section, source=None, lineno=None):
        msg = "Section " + repr(section) + " already exists"
        if source is not None:
            msg = "While reading from " + repr(source) + " [line " + ("%2d" % lineno) + "]: section " + repr(section) + " already exists"
        Error.__init__(self, msg)
        self.section = section
        self.source = source
        self.lineno = lineno


class DuplicateOptionError(Error):
    def __init__(self, section, option, source=None, lineno=None):
        msg = "Option " + repr(option) + " in section " + repr(section) + " already exists"
        if source is not None:
            msg = "While reading from " + repr(source) + " [line " + ("%2d" % lineno) + "]: option " + repr(option) + " in section " + repr(section) + " already exists"
        Error.__init__(self, msg)
        self.section = section
        self.option = option
        self.source = source
        self.lineno = lineno


class NoOptionError(Error):
    def __init__(self, option, section):
        Error.__init__(self, "No option " + repr(option) + " in section: " + repr(section))
        self.option = option
        self.section = section


class InterpolationError(Error):
    def __init__(self, option, section, msg):
        Error.__init__(self, msg)
        self.option = option
        self.section = section


class InterpolationMissingOptionError(InterpolationError):
    def __init__(self, option, section, rawval, reference):
        msg = "Bad value substitution: option " + repr(option) + " in section " + repr(section) + " contains an interpolation key " + repr(reference) + " which is not a valid option name. Raw value: " + repr(rawval)
        InterpolationError.__init__(self, option, section, msg)
        self.reference = reference


class InterpolationSyntaxError(InterpolationError):
    pass


class InterpolationDepthError(InterpolationError):
    def __init__(self, option, section, rawval):
        msg = "Recursion limit exceeded in value substitution: option " + repr(option) + " in section " + repr(section) + " contains an interpolation key which cannot be substituted in " + str(MAX_INTERPOLATION_DEPTH) + " steps. Raw value: " + repr(rawval)
        InterpolationError.__init__(self, option, section, msg)


class ParsingError(Error):
    def __init__(self, source=None, filename=None):
        if filename and source:
            raise ValueError("Cannot specify both `filename' and `source'. Use `source'.")
        elif not filename and not source:
            raise ValueError("Required argument `source' not given.")
        elif filename:
            source = filename
        Error.__init__(self, "Source contains parsing errors: " + repr(source))
        self.source = source
        self.errors = []

    @property
    def filename(self):
        return self.source

    def append(self, lineno, line):
        self.errors.append((lineno, line))
        self.message += "\n\t[line " + ("%2d" % lineno) + "]: " + repr(line)


class MissingSectionHeaderError(ParsingError):
    def __init__(self, filename, lineno, line):
        Error.__init__(self, "File contains no section headers.\nfile: " + repr(filename) + ", line: " + str(lineno) + "\n" + repr(line))
        self.source = filename
        self.errors = []
        self.lineno = lineno
        self.line = line


class Interpolation:
    def before_get(self, parser, section, option, value, defaults):
        return value

    def before_set(self, parser, section, option, value):
        return value

    def before_read(self, parser, section, option, value):
        return value

    def before_write(self, parser, section, option, value):
        return value


class BasicInterpolation(Interpolation):
    _KEYCRE = re.compile(r"%\(([^)]+)\)s")

    def before_get(self, parser, section, option, value, defaults):
        L = []
        self._interpolate_some(parser, option, L, value, section, defaults, 1)
        return "".join(L)

    def before_set(self, parser, section, option, value):
        tmp_value = value.replace("%%", "")
        tmp_value = self._KEYCRE.sub("", tmp_value)
        if "%" in tmp_value:
            raise ValueError("invalid interpolation syntax in " + repr(value) + " at position " + str(tmp_value.find("%")))
        return value

    def _interpolate_some(self, parser, option, accum, rest, section, map, depth):
        rawval = parser.get(section, option, raw=True, fallback=rest)
        if depth > MAX_INTERPOLATION_DEPTH:
            raise InterpolationDepthError(option, section, rawval)
        while rest:
            p = rest.find("%")
            if p < 0:
                accum.append(rest)
                return
            if p > 0:
                accum.append(rest[:p])
                rest = rest[p:]
            c = rest[1:2]
            if c == "%":
                accum.append("%")
                rest = rest[2:]
            elif c == "(":
                m = self._KEYCRE.match(rest)
                if m is None:
                    raise InterpolationSyntaxError(option, section, "bad interpolation variable reference " + repr(rest))
                var = parser.optionxform(m.group(1))
                rest = rest[m.end():]
                try:
                    v = map[var]
                except KeyError:
                    raise InterpolationMissingOptionError(option, section, rawval, var) from None
                if "%" in v:
                    self._interpolate_some(parser, option, accum, v, section, map, depth + 1)
                else:
                    accum.append(v)
            else:
                raise InterpolationSyntaxError(option, section, "'%' must be followed by '%' or '(', found: " + repr(rest))


class ExtendedInterpolation(Interpolation):
    _KEYCRE = re.compile(r"\$\{([^}]+)\}")

    def before_get(self, parser, section, option, value, defaults):
        L = []
        self._interpolate_some(parser, option, L, value, section, defaults, 1)
        return "".join(L)

    def before_set(self, parser, section, option, value):
        tmp_value = value.replace("$$", "")
        tmp_value = self._KEYCRE.sub("", tmp_value)
        if "$" in tmp_value:
            raise ValueError("invalid interpolation syntax in " + repr(value) + " at position " + str(tmp_value.find("$")))
        return value

    def _interpolate_some(self, parser, option, accum, rest, section, map, depth):
        rawval = parser.get(section, option, raw=True, fallback=rest)
        if depth > MAX_INTERPOLATION_DEPTH:
            raise InterpolationDepthError(option, section, rawval)
        while rest:
            p = rest.find("$")
            if p < 0:
                accum.append(rest)
                return
            if p > 0:
                accum.append(rest[:p])
                rest = rest[p:]
            c = rest[1:2]
            if c == "$":
                accum.append("$")
                rest = rest[2:]
            elif c == "{":
                m = self._KEYCRE.match(rest)
                if m is None:
                    raise InterpolationSyntaxError(option, section, "bad interpolation variable reference " + repr(rest))
                path = m.group(1).split(":")
                rest = rest[m.end():]
                sect = section
                opt = option
                try:
                    if len(path) == 1:
                        opt = parser.optionxform(path[0])
                        v = map[opt]
                    elif len(path) == 2:
                        sect = path[0]
                        opt = parser.optionxform(path[1])
                        v = parser.get(sect, opt, raw=True)
                    else:
                        raise InterpolationSyntaxError(option, section, "More than one ':' found: " + repr(rest))
                except (KeyError, NoSectionError, NoOptionError):
                    raise InterpolationMissingOptionError(option, section, rawval, ":".join(path)) from None
                if "$" in v:
                    self._interpolate_some(parser, opt, accum, v, sect, dict(parser.items(sect, raw=True)), depth + 1)
                else:
                    accum.append(v)
            else:
                raise InterpolationSyntaxError(option, section, "'$' must be followed by '$' or '{', found: " + repr(rest))


class RawConfigParser:
    SECTCRE = re.compile(_SECT_TMPL, re.VERBOSE)
    NONSPACECRE = re.compile(r"\S")
    BOOLEAN_STATES = {"1": True, "yes": True, "true": True, "on": True, "0": False, "no": False, "false": False, "off": False}

    def __init__(self, defaults=None, dict_type=dict, allow_no_value=False, *, delimiters=("=", ":"), comment_prefixes=("#", ";"), inline_comment_prefixes=None, strict=True, empty_lines_in_values=True, default_section=DEFAULTSECT, interpolation=_UNSET, converters=_UNSET):
        self._dict = dict_type
        self._sections = self._dict()
        self._defaults = self._dict()
        self._converters = {}
        self._proxies = self._dict()
        self._proxies[default_section] = SectionProxy(self, default_section)
        self._delimiters = tuple(delimiters)
        if delimiters == ("=", ":"):
            self._optcre = re.compile(_OPT_NV_TMPL.format(delim="=|:") if allow_no_value else _OPT_TMPL.format(delim="=|:"), re.VERBOSE)
        else:
            d = "|".join([re.escape(x) for x in delimiters])
            if allow_no_value:
                self._optcre = re.compile(_OPT_NV_TMPL.format(delim=d), re.VERBOSE)
            else:
                self._optcre = re.compile(_OPT_TMPL.format(delim=d), re.VERBOSE)
        self._comment_prefixes = tuple(comment_prefixes or ())
        self._inline_comment_prefixes = tuple(inline_comment_prefixes or ())
        self._strict = strict
        self._allow_no_value = allow_no_value
        self._empty_lines_in_values = empty_lines_in_values
        self.default_section = default_section
        if interpolation is _UNSET:
            self._interpolation = self._default_interpolation()
        elif interpolation is None:
            self._interpolation = Interpolation()
        else:
            self._interpolation = interpolation
        if defaults:
            self._read_defaults(defaults)

    def _default_interpolation(self):
        return _obj(Interpolation())

    def defaults(self):
        return self._defaults

    def sections(self):
        return list(self._sections.keys())

    def add_section(self, section):
        if section == self.default_section:
            raise ValueError("Invalid section name: " + repr(section))
        if section in self._sections:
            raise DuplicateSectionError(section)
        self._sections[section] = self._dict()
        self._proxies[section] = SectionProxy(self, section)

    def has_section(self, section):
        return section in self._sections

    def options(self, section):
        try:
            opts = self._sections[section].copy()
        except KeyError:
            raise NoSectionError(section) from None
        opts.update(self._defaults)
        return list(opts.keys())

    def read(self, filenames, encoding=None):
        if isinstance(filenames, str):
            filenames = [filenames]
        read_ok = []
        for filename in filenames:
            try:
                fp = open(filename, encoding=encoding)
            except OSError:
                continue
            with fp:
                self._read(fp, filename)
            read_ok.append(filename)
        return read_ok

    def read_file(self, f, source=None):
        if source is None:
            source = "<???>"
        self._read(f, source)

    def read_string(self, string, source="<string>"):
        sfile = _lines(string)
        self._read(sfile, source)

    def read_dict(self, dictionary, source="<dict>"):
        elements_added = set()
        for section, keys in dictionary.items():
            section = str(section)
            try:
                self.add_section(section)
            except (DuplicateSectionError, ValueError):
                if self._strict and section in elements_added:
                    raise
            elements_added.add(section)
            for key, value in keys.items():
                key = self.optionxform(str(key))
                if value is not None:
                    value = str(value)
                if self._strict and (section, key) in elements_added:
                    raise DuplicateOptionError(section, key, source)
                elements_added.add((section, key))
                self.set(section, key, value)

    def get(self, section, option, *, raw=False, vars=None, fallback=_UNSET):
        try:
            d = self._unify_values(section, vars)
        except NoSectionError:
            if fallback is _UNSET:
                raise
            return fallback
        option = self.optionxform(option)
        try:
            value = d[option]
        except KeyError:
            if fallback is _UNSET:
                raise NoOptionError(option, section) from None
            return fallback
        if raw or value is None:
            return value
        return self._interpolation.before_get(self, section, option, value, d)

    def _get(self, section, conv, option, **kwargs):
        return conv(self.get(section, option, **kwargs))

    def _get_conv(self, section, option, conv, *, raw=False, vars=None, fallback=_UNSET, **kwargs):
        try:
            return self._get(section, conv, option, raw=raw, vars=vars, **kwargs)
        except (NoSectionError, NoOptionError):
            if fallback is _UNSET:
                raise
            return fallback

    def getint(self, section, option, *, raw=False, vars=None, fallback=_UNSET, **kwargs):
        return self._get_conv(section, option, int, raw=raw, vars=vars, fallback=fallback, **kwargs)

    def getfloat(self, section, option, *, raw=False, vars=None, fallback=_UNSET, **kwargs):
        return self._get_conv(section, option, float, raw=raw, vars=vars, fallback=fallback, **kwargs)

    def getboolean(self, section, option, *, raw=False, vars=None, fallback=_UNSET, **kwargs):
        return self._get_conv(section, option, self._convert_to_boolean, raw=raw, vars=vars, fallback=fallback, **kwargs)

    def items(self, section=_UNSET, raw=False, vars=None):
        if section is _UNSET:
            return [(sect, self[sect]) for sect in self.keys()]
        d = self._defaults.copy()
        try:
            d.update(self._sections[section])
        except KeyError:
            if section != self.default_section:
                raise NoSectionError(section) from None
        orig_keys = list(d.keys())
        if vars:
            for key, value in vars.items():
                d[self.optionxform(key)] = value
        value_getter = lambda option: self._interpolation.before_get(self, section, option, d[option], d)
        if raw:
            value_getter = lambda option: d[option]
        return [(option, value_getter(option)) for option in orig_keys]

    def popitem(self):
        for key in self.sections():
            value = self[key]
            del self[key]
            return key, value
        raise KeyError

    def optionxform(self, optionstr):
        return optionstr.lower()

    def has_option(self, section, option):
        if not section or section == self.default_section:
            option = self.optionxform(option)
            return option in self._defaults
        elif section not in self._sections:
            return False
        else:
            option = self.optionxform(option)
            return option in self._sections[section] or option in self._defaults

    def set(self, section, option, value=None):
        if value:
            value = self._interpolation.before_set(self, section, option, value)
        if not section or section == self.default_section:
            sectdict = self._defaults
        else:
            try:
                sectdict = self._sections[section]
            except KeyError:
                raise NoSectionError(section) from None
        sectdict[self.optionxform(option)] = value

    def write(self, fp, space_around_delimiters=True):
        if space_around_delimiters:
            d = " {} ".format(self._delimiters[0])
        else:
            d = self._delimiters[0]
        if self._defaults:
            self._write_section(fp, self.default_section, self._defaults.items(), d)
        for section in self._sections:
            self._write_section(fp, section, self._sections[section].items(), d)

    def _write_section(self, fp, section_name, section_items, delimiter):
        fp.write("[" + str(section_name) + "]\n")
        for key, value in section_items:
            value = self._interpolation.before_write(self, section_name, key, value)
            if value is not None or not self._allow_no_value:
                value = delimiter + str(value).replace("\n", "\n\t")
            else:
                value = ""
            fp.write(str(key) + value + "\n")
        fp.write("\n")

    def remove_option(self, section, option):
        if not section or section == self.default_section:
            sectdict = self._defaults
        else:
            try:
                sectdict = self._sections[section]
            except KeyError:
                raise NoSectionError(section) from None
        option = self.optionxform(option)
        existed = option in sectdict
        if existed:
            del sectdict[option]
        return existed

    def remove_section(self, section):
        existed = section in self._sections
        if existed:
            del self._sections[section]
            del self._proxies[section]
        return existed

    def __getitem__(self, key):
        if key != self.default_section and not self.has_section(key):
            raise KeyError(key)
        return self._proxies[key]

    def __setitem__(self, key, value):
        if key in self and self[key] is value:
            return
        if key == self.default_section:
            self._defaults.clear()
        elif key in self._sections:
            self._sections[key].clear()
        self.read_dict({key: value})

    def __delitem__(self, key):
        if key == self.default_section:
            raise ValueError("Cannot remove the default section.")
        if not self.has_section(key):
            raise KeyError(key)
        self.remove_section(key)

    def __contains__(self, key):
        return key == self.default_section or self.has_section(key)

    def __len__(self):
        return len(self._sections) + 1

    def __iter__(self):
        return iter([self.default_section] + list(self._sections.keys()))

    def keys(self):
        return [self.default_section] + list(self._sections.keys())

    def values(self):
        return [self[k] for k in self.keys()]

    def _read(self, fp, fpname):
        elements_added = set()
        cursect = None
        sectname = None
        optname = None
        lineno = 0
        indent_level = 0
        e = None
        for line in fp:
            lineno += 1
            comment_start = sys.maxsize
            inline_prefixes = {}
            for p in self._inline_comment_prefixes:
                inline_prefixes[p] = -1
            while comment_start == sys.maxsize and inline_prefixes:
                next_prefixes = {}
                for prefix, index in inline_prefixes.items():
                    index = line.find(prefix, index + 1)
                    if index == -1:
                        continue
                    next_prefixes[prefix] = index
                    if index == 0 or (index > 0 and line[index - 1].isspace()):
                        comment_start = min(comment_start, index)
                inline_prefixes = next_prefixes
            for prefix in self._comment_prefixes:
                if line.strip().startswith(prefix):
                    comment_start = 0
                    break
            if comment_start == sys.maxsize:
                value = line.strip()
                has_comment = False
            else:
                value = line[:comment_start].strip()
                has_comment = True
            if not value:
                if self._empty_lines_in_values:
                    if not has_comment and cursect is not None and optname and cursect[optname] is not None:
                        cursect[optname].append("")
                else:
                    indent_level = sys.maxsize
                continue
            first_nonspace = self.NONSPACECRE.search(line)
            cur_indent_level = first_nonspace.start() if first_nonspace else 0
            if cursect is not None and optname and cur_indent_level > indent_level:
                cursect[optname].append(value)
            else:
                indent_level = cur_indent_level
                mo = self.SECTCRE.match(value)
                if mo:
                    sectname = mo.group("header")
                    if sectname in self._sections:
                        if self._strict and sectname in elements_added:
                            raise DuplicateSectionError(sectname, fpname, lineno)
                        cursect = self._sections[sectname]
                        elements_added.add(sectname)
                    elif sectname == self.default_section:
                        cursect = self._defaults
                    else:
                        cursect = self._dict()
                        self._sections[sectname] = cursect
                        self._proxies[sectname] = SectionProxy(self, sectname)
                        elements_added.add(sectname)
                    optname = None
                elif cursect is None:
                    raise MissingSectionHeaderError(fpname, lineno, line)
                else:
                    mo = self._optcre.match(value)
                    if mo:
                        optname = mo.group("option")
                        optval = mo.group("value")
                        if not optname:
                            e = self._handle_error(e, fpname, lineno, line)
                        optname = self.optionxform(optname.rstrip())
                        if self._strict and (sectname, optname) in elements_added:
                            raise DuplicateOptionError(sectname, optname, fpname, lineno)
                        elements_added.add((sectname, optname))
                        if optval is not None:
                            optval = optval.strip()
                            cursect[optname] = [optval]
                        else:
                            cursect[optname] = None
                    else:
                        e = self._handle_error(e, fpname, lineno, line)
        self._join_multiline_values()
        if e:
            raise e

    def _join_multiline_values(self):
        for name in list(self._defaults.keys()):
            val = self._defaults[name]
            if isinstance(val, list):
                val = "\n".join(val).rstrip()
            self._defaults[name] = self._interpolation.before_read(self, self.default_section, name, val)
        for section in list(self._sections.keys()):
            options = self._sections[section]
            for name in list(options.keys()):
                val = options[name]
                if isinstance(val, list):
                    val = "\n".join(val).rstrip()
                options[name] = self._interpolation.before_read(self, section, name, val)

    def _read_defaults(self, defaults):
        self.read_dict({self.default_section: defaults})

    def _handle_error(self, exc, fpname, lineno, line):
        if not exc:
            exc = ParsingError(fpname)
        exc.append(lineno, line)
        return exc

    def _unify_values(self, section, vars):
        sectiondict = {}
        try:
            sectiondict = self._sections[section]
        except KeyError:
            if section != self.default_section:
                raise NoSectionError(section) from None
        vardict = {}
        if vars:
            for key, value in vars.items():
                if value is not None:
                    value = str(value)
                vardict[self.optionxform(key)] = value
        return ChainMap(vardict, sectiondict, self._defaults)

    def _convert_to_boolean(self, value):
        if value.lower() not in self.BOOLEAN_STATES:
            raise ValueError("Not a boolean: " + value)
        return self.BOOLEAN_STATES[value.lower()]

    def _validate_value_types(self, *, section="", option="", value=""):
        if not isinstance(section, str):
            raise TypeError("section names must be strings")
        if not isinstance(option, str):
            raise TypeError("option keys must be strings")
        if not self._allow_no_value or value:
            if not isinstance(value, str):
                raise TypeError("option values must be strings")


def _lines(string):
    out = []
    for line in string.split("\n"):
        out.append(line + "\n")
    if out and out[-1] == "\n" and string.endswith("\n"):
        out.pop()
    elif out:
        out[-1] = out[-1][:-1]
    return out


class ConfigParser(RawConfigParser):
    def _default_interpolation(self):
        return _obj(BasicInterpolation())

    def set(self, section, option, value=None):
        self._validate_value_types(option=option, value=value)
        RawConfigParser.set(self, section, option, value)

    def add_section(self, section):
        self._validate_value_types(section=section)
        RawConfigParser.add_section(self, section)

    def _read_defaults(self, defaults):
        hold_interpolation = self._interpolation
        self._interpolation = Interpolation()
        try:
            self.read_dict({self.default_section: defaults})
        finally:
            self._interpolation = hold_interpolation


class SectionProxy:
    def __init__(self, parser, name):
        self._parser = parser
        self._name = name

    def __repr__(self):
        return "<Section: " + self._name + ">"

    def __getitem__(self, key):
        if not self._parser.has_option(self._name, key):
            raise KeyError(key)
        return self._parser.get(self._name, key)

    def __setitem__(self, key, value):
        self._parser._validate_value_types(option=key, value=value)
        self._parser.set(self._name, key, value)

    def __delitem__(self, key):
        if not (self._parser.has_option(self._name, key) and self._parser.remove_option(self._name, key)):
            raise KeyError(key)

    def __contains__(self, key):
        return self._parser.has_option(self._name, key)

    def __len__(self):
        return len(self._options())

    def __iter__(self):
        return iter(self._options())

    def _options(self):
        if self._name != self._parser.default_section:
            return self._parser.options(self._name)
        else:
            return list(self._parser.defaults().keys())

    @property
    def parser(self):
        return self._parser

    @property
    def name(self):
        return self._name

    def get(self, option, fallback=None, *, raw=False, vars=None, _impl=None, **kwargs):
        return self._parser.get(self._name, option, raw=raw, vars=vars, fallback=fallback)

    def getint(self, option, fallback=None, *, raw=False, vars=None):
        return self._parser.getint(self._name, option, raw=raw, vars=vars, fallback=fallback)

    def getfloat(self, option, fallback=None, *, raw=False, vars=None):
        return self._parser.getfloat(self._name, option, raw=raw, vars=vars, fallback=fallback)

    def getboolean(self, option, fallback=None, *, raw=False, vars=None):
        return self._parser.getboolean(self._name, option, raw=raw, vars=vars, fallback=fallback)

    def keys(self):
        return self._options()

    def values(self):
        return [self[k] for k in self._options()]

    def items(self):
        return [(k, self[k]) for k in self._options()]
