words: list[str] = ["a", "bb"]
empty: list[str] = []


def bad(w: str) -> int:
    return 10 // (len(w) - 1)


try:
    print(sorted(words, key=bad))
except ZeroDivisionError as e:
    print("ZeroDivisionError:", e)
try:
    print(max(empty, key=len))
except ValueError as e:
    print("ValueError:", e)
