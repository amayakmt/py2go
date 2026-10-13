import ns.top
import ns.inner.deep
from ns.inner.deep import VALUE
print(ns.top.call(), ns.inner.deep.VALUE, VALUE)
