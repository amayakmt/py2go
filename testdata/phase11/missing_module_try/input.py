try:
    import nothere
except ImportError as e:
    print(type(e).__name__, e)
try:
    from nothere.sub import thing
except ModuleNotFoundError as e:
    print("MNF", e)
try:
    import fastlib
    HAVE = True
except ImportError:
    HAVE = False
print(HAVE)
import present
try:
    from present import absent
except ImportError as e:
    print("ImportError", str(e).split(" (")[0])
print(present.ok)
