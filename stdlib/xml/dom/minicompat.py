"""Python version compatibility support for minidom.

This module contains internal implementation details and
should not be imported; use xml.dom.minidom instead.
"""
__all__ = ['NodeList', 'EmptyNodeList', 'StringTypes', 'defproperty']
import xml.dom
StringTypes = (str,)

class NodeList(list):

    def item(self, index):
        if 0 <= index < len(self):
            return self[index]

    def _get_length(self):
        return len(self)

    def _set_length(self, value):
        raise xml.dom.NoModificationAllowedErr("attempt to modify read-only attribute 'length'")
    length = property(_get_length, _set_length, doc='The number of nodes in the NodeList.')

    def __setstate__(self, state):
        if state is None:
            state = []
        self.clear()
        self.extend(state)

class EmptyNodeList(tuple):

    def __add__(self, other):
        NL = NodeList()
        NL.extend(other)
        return NL

    def __radd__(self, other):
        NL = NodeList()
        NL.extend(other)
        return NL

    def item(self, index):
        return None

    def _get_length(self):
        return 0

    def _set_length(self, value):
        raise xml.dom.NoModificationAllowedErr("attempt to modify read-only attribute 'length'")
    length = property(_get_length, _set_length, doc='The number of nodes in the NodeList.')

def defproperty(klass, name, doc):
    """The linker turns the module-level calls of defproperty into
    properties of the class, since a class cannot get a property at run time."""
    raise NotImplementedError('defproperty must be called at the top level of the module that defines the class')
