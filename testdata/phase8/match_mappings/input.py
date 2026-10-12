def handle(msg: dict[str, int]) -> str:
    match msg:
        case {"type": 1, "id": ident}:
            return "type one " + str(ident)
        case {"type": t, **rest}:
            return f"type {t} rest {rest}"
        case {}:
            return "empty-or-other"


print(handle({"type": 1, "id": 7}))
print(handle({"type": 2, "a": 1, "b": 2}))
print(handle({}))
print(handle({"zzz": 1}))
