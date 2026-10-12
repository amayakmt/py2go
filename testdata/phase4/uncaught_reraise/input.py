try:
    raise KeyError("missing")
except KeyError as e:
    print("log:", e)
    raise
