from enum import IntEnum, StrEnum, Flag, IntFlag, auto
class Pri(IntEnum):
    LOW = 1
    MID = 5
    HIGH = 10
class Mode(StrEnum):
    READ = "r"
    WRITE = "w"
    AUTO = auto()
class Perm(Flag):
    R = auto()
    W = auto()
    X = auto()
class Bits(IntFlag):
    A = 1
    B = 2
    C = 4
print(Pri.LOW, repr(Pri.MID), Pri.HIGH + 1, Pri.LOW < Pri.HIGH, Pri.LOW == 1, 5 == Pri.MID, int(Pri.HIGH), Pri(5), sorted([Pri.HIGH, Pri.LOW, Pri.MID]), max(Pri), Pri.MID * 2, [p.value for p in Pri], f"{Pri.HIGH:>4}", hash(Pri.LOW) == hash(1), {Pri.LOW: "a"}[1] if False else "skip")
print(Mode.READ, repr(Mode.WRITE), Mode.AUTO.value, Mode.READ == "r", Mode.READ.upper(), Mode("w"), f"{Mode.READ}", Mode.READ + "x", len(Mode.READ), list(Mode), "r" in Mode.READ)
p = Perm.R | Perm.W
print(p, repr(p), p.value, Perm.R in p, Perm.X in p, p & Perm.R, p & Perm.X, repr(p ^ Perm.R), bool(Perm.R & Perm.W), Perm(5), Perm.R | Perm.W | Perm.X, list(Perm), list(p), ~Perm.R, Perm.R == Perm.R, Perm.R != Perm.W)
b = Bits.A | Bits.C
print(b, repr(b), int(b), b == 5, b & 4, Bits(3), Bits.A | 2, 4 | Bits.A, b.value, Bits.B in b, [x for x in Bits], bool(b & Bits.B))
for bad in (lambda: Pri(2), lambda: Mode("x"), lambda: Perm(16), lambda: Pri["NOPE"]):
    try:
        print(bad())
    except (ValueError, KeyError) as e:
        print(type(e).__name__, e)
