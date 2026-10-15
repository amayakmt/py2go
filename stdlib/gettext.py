"""Internationalization and localization support (reduced).

No message catalogs are read: every translation is the identity, as in a
program run without a locale directory."""
__all__ = ['NullTranslations', 'translation', 'install', 'textdomain', 'bindtextdomain', 'dgettext', 'dngettext', 'gettext', 'ngettext', 'pgettext', 'npgettext', 'find', 'GNUTranslations']

_current_domain = 'messages'


class NullTranslations:

    def __init__(self, fp=None):
        self._fallback = None

    def gettext(self, message):
        return message

    def ngettext(self, msgid1, msgid2, n):
        if n == 1:
            return msgid1
        return msgid2

    def pgettext(self, context, message):
        return message

    def npgettext(self, context, msgid1, msgid2, n):
        if n == 1:
            return msgid1
        return msgid2

    def add_fallback(self, fallback):
        self._fallback = fallback

    def install(self, names=None):
        pass


class GNUTranslations(NullTranslations):
    pass


def find(domain, localedir=None, languages=None, all=False):
    if all:
        return []
    return None


def translation(domain, localedir=None, languages=None, class_=None, fallback=False):
    if not fallback:
        raise FileNotFoundError(2, 'No translation file found for domain', domain)
    return NullTranslations()


def install(domain, localedir=None, *, names=None):
    pass


def textdomain(domain=None):
    global _current_domain
    if domain is not None:
        _current_domain = domain
    return _current_domain


def bindtextdomain(domain, localedir=None):
    return localedir


def dgettext(domain, message):
    return message


def dngettext(domain, msgid1, msgid2, n):
    if n == 1:
        return msgid1
    return msgid2


def gettext(message):
    return message


def ngettext(msgid1, msgid2, n):
    if n == 1:
        return msgid1
    return msgid2


def pgettext(context, message):
    return message


def npgettext(context, msgid1, msgid2, n):
    if n == 1:
        return msgid1
    return msgid2
