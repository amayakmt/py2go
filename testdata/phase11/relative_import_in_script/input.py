try:
    from . import something
except ImportError as e:
    print("ImportError:", e)
print("ok")
