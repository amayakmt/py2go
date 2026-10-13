import atexit
import threading
import time
import sys


def bye(name, punct="."):
    print("bye " + name + punct)


atexit.register(bye, "first")
atexit.register(bye, "second", punct="!")
gone = atexit.register(bye, "never")
atexit.unregister(bye)
atexit.register(bye, "third")


def late():
    time.sleep(0.05)
    print("non-daemon thread finished")


threading.Thread(target=late).start()


def daemon():
    time.sleep(5)
    print("never printed")


threading.Thread(target=daemon, daemon=True).start()
print("main done")
sys.stdout.flush()
