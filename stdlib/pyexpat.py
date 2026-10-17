"""Python wrapper for the Expat parser, over the XML parser of the runtime.

The parser reads the whole document when Parse is called with isfinal true
(or at ParseFile), then calls the handlers in order: data fed in several
Parse calls is buffered until the end. The events, error codes, messages,
lines and columns are expat's."""

EXPAT_VERSION = 'expat_2.8.4'
version_info = (2, 8, 4)
native_encoding = 'UTF-8'
features = [('sizeof(XML_Char)', 1), ('sizeof(XML_LChar)', 1), ('XML_DTD', 0), ('XML_CONTEXT_BYTES', 1024), ('XML_NS', 0)]
XML_PARAM_ENTITY_PARSING_NEVER = 0
XML_PARAM_ENTITY_PARSING_UNLESS_STANDALONE = 1
XML_PARAM_ENTITY_PARSING_ALWAYS = 2


def _obj(x: object) -> object:
    return x


class ExpatError(Exception):

    def __init__(self, *args):
        self.args = args
        self.code = _obj(None)
        self.lineno = _obj(None)
        self.offset = _obj(None)


error = ExpatError


class errors:
    """Constants and messages of the errors expat reports."""
    XML_ERROR_ABORTED = 'parsing aborted'
    XML_ERROR_AMPLIFICATION_LIMIT_BREACH = 'limit on input amplification factor (from DTD and entities) breached'
    XML_ERROR_ASYNC_ENTITY = 'asynchronous entity'
    XML_ERROR_ATTRIBUTE_EXTERNAL_ENTITY_REF = 'reference to external entity in attribute'
    XML_ERROR_BAD_CHAR_REF = 'reference to invalid character number'
    XML_ERROR_BINARY_ENTITY_REF = 'reference to binary entity'
    XML_ERROR_CANT_CHANGE_FEATURE_ONCE_PARSING = 'cannot change setting once parsing has begun'
    XML_ERROR_DUPLICATE_ATTRIBUTE = 'duplicate attribute'
    XML_ERROR_ENTITY_DECLARED_IN_PE = 'entity declared in parameter entity'
    XML_ERROR_EXTERNAL_ENTITY_HANDLING = 'error in processing external entity reference'
    XML_ERROR_FEATURE_REQUIRES_XML_DTD = 'requested feature requires XML_DTD support in Expat'
    XML_ERROR_FINISHED = 'parsing finished'
    XML_ERROR_INCOMPLETE_PE = 'incomplete markup in parameter entity'
    XML_ERROR_INCORRECT_ENCODING = 'encoding specified in XML declaration is incorrect'
    XML_ERROR_INVALID_ARGUMENT = 'invalid argument'
    XML_ERROR_INVALID_TOKEN = 'not well-formed (invalid token)'
    XML_ERROR_JUNK_AFTER_DOC_ELEMENT = 'junk after document element'
    XML_ERROR_MISPLACED_XML_PI = 'XML or text declaration not at start of entity'
    XML_ERROR_NOT_STANDALONE = 'document is not standalone'
    XML_ERROR_NOT_SUSPENDED = 'parser not suspended'
    XML_ERROR_NO_BUFFER = 'a successful prior call to function XML_GetBuffer is required'
    XML_ERROR_NO_ELEMENTS = 'no element found'
    XML_ERROR_NO_MEMORY = 'out of memory'
    XML_ERROR_PARAM_ENTITY_REF = 'illegal parameter entity reference'
    XML_ERROR_PARTIAL_CHAR = 'partial character'
    XML_ERROR_PUBLICID = 'illegal character(s) in public id'
    XML_ERROR_RECURSIVE_ENTITY_REF = 'recursive entity reference'
    XML_ERROR_RESERVED_NAMESPACE_URI = 'prefix must not be bound to one of the reserved namespace names'
    XML_ERROR_RESERVED_PREFIX_XML = 'reserved prefix (xml) must not be undeclared or bound to another namespace name'
    XML_ERROR_RESERVED_PREFIX_XMLNS = 'reserved prefix (xmlns) must not be declared or undeclared'
    XML_ERROR_SUSPENDED = 'parser suspended'
    XML_ERROR_SUSPEND_PE = 'cannot suspend in external parameter entity'
    XML_ERROR_SYNTAX = 'syntax error'
    XML_ERROR_TAG_MISMATCH = 'mismatched tag'
    XML_ERROR_TEXT_DECL = 'text declaration not well-formed'
    XML_ERROR_UNBOUND_PREFIX = 'unbound prefix'
    XML_ERROR_UNCLOSED_CDATA_SECTION = 'unclosed CDATA section'
    XML_ERROR_UNCLOSED_TOKEN = 'unclosed token'
    XML_ERROR_UNDECLARING_PREFIX = 'must not undeclare prefix'
    XML_ERROR_UNDEFINED_ENTITY = 'undefined entity'
    XML_ERROR_UNEXPECTED_STATE = 'unexpected parser state - please send a bug report'
    XML_ERROR_UNKNOWN_ENCODING = 'unknown encoding'
    XML_ERROR_XML_DECL = 'XML declaration not well-formed'
    codes = {'parsing aborted': 35, 'limit on input amplification factor (from DTD and entities) breached': 43, 'asynchronous entity': 13, 'reference to external entity in attribute': 16, 'reference to invalid character number': 14, 'reference to binary entity': 15, 'cannot change setting once parsing has begun': 26, 'duplicate attribute': 8, 'entity declared in parameter entity': 24, 'error in processing external entity reference': 21, 'requested feature requires XML_DTD support in Expat': 25, 'parsing finished': 36, 'incomplete markup in parameter entity': 29, 'encoding specified in XML declaration is incorrect': 19, 'invalid argument': 41, 'not well-formed (invalid token)': 4, 'junk after document element': 9, 'XML or text declaration not at start of entity': 17, 'document is not standalone': 22, 'parser not suspended': 34, 'a successful prior call to function XML_GetBuffer is required': 42, 'no element found': 3, 'out of memory': 1, 'illegal parameter entity reference': 10, 'partial character': 6, 'illegal character(s) in public id': 32, 'recursive entity reference': 12, 'prefix must not be bound to one of the reserved namespace names': 40, 'reserved prefix (xml) must not be undeclared or bound to another namespace name': 38, 'reserved prefix (xmlns) must not be declared or undeclared': 39, 'parser suspended': 33, 'cannot suspend in external parameter entity': 37, 'syntax error': 2, 'mismatched tag': 7, 'text declaration not well-formed': 31, 'unbound prefix': 27, 'unclosed CDATA section': 20, 'unclosed token': 5, 'must not undeclare prefix': 28, 'undefined entity': 11, 'unexpected parser state - please send a bug report': 23, 'unknown encoding': 18, 'XML declaration not well-formed': 30}
    messages = _obj({})


for _msg in errors.codes:
    errors.messages[errors.codes[_msg]] = _msg


class model:
    XML_CQUANT_NONE = 0
    XML_CQUANT_OPT = 1
    XML_CQUANT_PLUS = 3
    XML_CQUANT_REP = 2
    XML_CTYPE_ANY = 2
    XML_CTYPE_CHOICE = 5
    XML_CTYPE_EMPTY = 1
    XML_CTYPE_MIXED = 3
    XML_CTYPE_NAME = 4
    XML_CTYPE_SEQ = 6


def ErrorString(code):
    return _native("pyexpat.error_string", code)


def _detect_encoding(data, declared):
    if declared:
        return declared
    if data[:3] == b'\xef\xbb\xbf':
        return 'utf-8-sig'
    if data[:2] == b'\xff\xfe' or data[:2] == b'\xfe\xff':
        return 'utf-16'
    if data[:5] == b'<?xml':
        end = data.find(b'?>')
        head = data[:end if end >= 0 else 200]
        i = head.find(b'encoding')
        if i >= 0:
            rest = head[i + 8:].lstrip(b' \t\r\n')
            if rest[:1] == b'=':
                rest = rest[1:].lstrip(b' \t\r\n')
                q = rest[:1]
                if q == b'"' or q == b"'":
                    j = rest.find(q, 1)
                    if j > 0:
                        return rest[1:j].decode('ascii')
    return 'utf-8'


class xmlparser:

    def __init__(self, encoding=None, namespace_separator=None):
        self._encoding = _obj(encoding)
        self._sep = _obj(namespace_separator)
        self._chunks = _obj([])
        self._finished = False
        self._base = _obj(None)
        self.buffer_text = _obj(False)
        self.buffer_size = 8192
        self.buffer_used = 0
        self.ordered_attributes = _obj(False)
        self.specified_attributes = _obj(False)
        self.namespace_prefixes = _obj(False)
        self.intern = _obj({})
        self.ErrorCode = 0
        self.ErrorLineNumber = 1
        self.ErrorColumnNumber = 0
        self.ErrorByteIndex = -1
        self.CurrentLineNumber = 1
        self.CurrentColumnNumber = 0
        self.CurrentByteIndex = 0
        self.StartElementHandler = _obj(None)
        self.EndElementHandler = _obj(None)
        self.CharacterDataHandler = _obj(None)
        self.ProcessingInstructionHandler = _obj(None)
        self.CommentHandler = _obj(None)
        self.StartNamespaceDeclHandler = _obj(None)
        self.EndNamespaceDeclHandler = _obj(None)
        self.StartCdataSectionHandler = _obj(None)
        self.EndCdataSectionHandler = _obj(None)
        self.XmlDeclHandler = _obj(None)
        self.StartDoctypeDeclHandler = _obj(None)
        self.EndDoctypeDeclHandler = _obj(None)
        self.DefaultHandler = _obj(None)
        self.DefaultHandlerExpand = _obj(None)
        self.NotStandaloneHandler = _obj(None)
        self.ExternalEntityRefHandler = _obj(None)
        self.ElementDeclHandler = _obj(None)
        self.AttlistDeclHandler = _obj(None)
        self.EntityDeclHandler = _obj(None)
        self.NotationDeclHandler = _obj(None)
        self.UnparsedEntityDeclHandler = _obj(None)
        self.SkippedEntityHandler = _obj(None)

    def Parse(self, data, isfinal=False):
        if self._finished:
            err = ExpatError('parsing finished: line %d, column %d' % (self.ErrorLineNumber, self.ErrorColumnNumber))
            err.code = 36
            raise err
        if data:
            self._chunks.append(data)
        if not isfinal:
            return 1
        self._finished = True
        self._run()
        return 1

    def ParseFile(self, file):
        while True:
            buf = file.read(65536)
            if not buf:
                break
            self._chunks.append(buf)
        self._finished = True
        self._run()
        return 1

    def _text(self):
        if not self._chunks:
            return ''
        if isinstance(self._chunks[0], str):
            return ''.join([c if isinstance(c, str) else c.decode('utf-8') for c in self._chunks])
        data = b''.join([c if isinstance(c, bytes) else c.encode('utf-8') for c in self._chunks])
        enc = _detect_encoding(data, self._encoding)
        try:
            return data.decode(enc)
        except UnicodeDecodeError:
            self._raise(4, 'not well-formed (invalid token)', 1, 0)
        except LookupError:
            self._raise(18, 'unknown encoding', 1, 0)
        return ''

    def _raise(self, code, message, line, col):
        self.ErrorCode = code
        self.ErrorLineNumber = line
        self.ErrorColumnNumber = col
        err = ExpatError('%s: line %d, column %d' % (message, line, col))
        err.code = code
        err.lineno = line
        err.offset = col
        raise err

    def _run(self):
        text = self._text()
        result = _native("pyexpat.parse", text, self._sep)
        events = result[0]
        problem = result[1]
        pending = []
        for ev in events:
            kind = ev[0]
            if kind == 'data':
                if self.CharacterDataHandler is not None:
                    if self.buffer_text:
                        pending.append(ev[1])
                    else:
                        self.CharacterDataHandler(ev[1])
                continue
            if pending:
                self.CharacterDataHandler(''.join(pending))
                pending = []
            if kind == 'start':
                if self.StartElementHandler is not None:
                    flat = ev[2]
                    if self.ordered_attributes:
                        self.StartElementHandler(ev[1], list(flat))
                    else:
                        attrs = {}
                        i = 0
                        while i < len(flat):
                            attrs[flat[i]] = flat[i + 1]
                            i += 2
                        self.StartElementHandler(ev[1], attrs)
            elif kind == 'end':
                if self.EndElementHandler is not None:
                    self.EndElementHandler(ev[1])
            elif kind == 'comment':
                if self.CommentHandler is not None:
                    self.CommentHandler(ev[1])
            elif kind == 'pi':
                if self.ProcessingInstructionHandler is not None:
                    self.ProcessingInstructionHandler(ev[1], ev[2])
            elif kind == 'startns':
                if self.StartNamespaceDeclHandler is not None:
                    self.StartNamespaceDeclHandler(ev[1], ev[2])
            elif kind == 'endns':
                if self.EndNamespaceDeclHandler is not None:
                    self.EndNamespaceDeclHandler(ev[1])
            elif kind == 'startcdata':
                if self.StartCdataSectionHandler is not None:
                    self.StartCdataSectionHandler()
            elif kind == 'endcdata':
                if self.EndCdataSectionHandler is not None:
                    self.EndCdataSectionHandler()
            elif kind == 'xmldecl':
                if self.XmlDeclHandler is not None:
                    self.XmlDeclHandler(ev[1], ev[2], ev[3])
            elif kind == 'startdoctype':
                if self.StartDoctypeDeclHandler is not None:
                    self.StartDoctypeDeclHandler(ev[1], ev[2], ev[3], 1 if ev[4] else 0)
            elif kind == 'enddoctype':
                if self.EndDoctypeDeclHandler is not None:
                    self.EndDoctypeDeclHandler()
        if pending:
            self.CharacterDataHandler(''.join(pending))
        if problem is not None:
            self._raise(problem[0], problem[1], problem[2], problem[3])

    def GetReparseDeferralEnabled(self):
        return True

    def SetReparseDeferralEnabled(self, enabled):
        pass

    def SetParamEntityParsing(self, flag):
        return 1

    def UseForeignDTD(self, flag=True):
        pass

    def SetBase(self, base):
        self._base = base

    def GetBase(self):
        return self._base

    def GetInputContext(self):
        return None

    def ExternalEntityParserCreate(self, context, encoding=None):
        return xmlparser(encoding, self._sep)


XMLParserType = xmlparser


def ParserCreate(encoding=None, namespace_separator=None, intern=None):
    if namespace_separator is not None and len(namespace_separator) > 1:
        raise ValueError('namespace_separator must be at most one character, omitted, or None')
    return xmlparser(encoding, namespace_separator)
