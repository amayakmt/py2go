import os
import sys
import math


def describe(m):
    return m.__name__, hasattr(m, "sqrt"), getattr(m, "nope", "dflt")


print(describe(math), hasattr(os, "getcwd"), hasattr(os, "no_such"))
print(getattr(os, "sep"), os.path is not None)
p = os.path
print(p.join("a", "b"))
print(type(math).__name__ if False else "skip")
print(getattr(sys, "platform"))
import json


def pick(mod, name):
    return getattr(mod, name)(0) if hasattr(mod, name) else None


print(pick(math, "cos"), pick(math, "nothing"), pick(json, "dumps"))
mods = [math, json]
print([m.__name__ for m in mods])
print("sqrt" in dir(math), "no_such_name" in dir(math))
