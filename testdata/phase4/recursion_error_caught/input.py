def forever(n: int) -> int:
    return forever(n + 1) + 1


try:
    forever(0)
except RecursionError as e:
    print("caught:", e)
try:
    forever(0)
except RuntimeError as e:
    print("also a RuntimeError:", type(e).__name__)
print("still running")
