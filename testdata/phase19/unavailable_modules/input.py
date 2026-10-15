import sys

try:
    import msvcrt
except ImportError as e:
    print(type(e).__name__, e)

try:
    import winreg
except ModuleNotFoundError as e:
    print("winreg missing")

try:
    from winsound import Beep
except ImportError as e:
    print(e)

try:
    import _winapi
except ImportError as e:
    print(e)

try:
    import nt
except ImportError:
    nt = None
print(nt, sys.platform != "win32")
