import dbm
import os
import shelve
import tempfile

d = tempfile.mkdtemp()
base = os.path.join(d, "db")
with dbm.open(base, "c") as db:
    db["alpha"] = "one"
    db[b"beta"] = b"two"
    db["alpha"] = "uno"
    print(db["alpha"], db[b"beta"], sorted(db.keys()), len(db), "alpha" in db, "zzz" in db, db.get("zzz", b"d"))
    del db["beta"]
    print(sorted(db.keys()))
    try:
        db["gone"]
    except KeyError as e:
        print("KeyError", e)
with dbm.open(base, "r") as db:
    print(db["alpha"], list(db.keys()))
try:
    dbm.open(os.path.join(d, "missing"), "r")
except Exception as e:
    print(type(e).__name__, e)

path = os.path.join(d, "shelf")
with shelve.open(path) as sh:
    sh["list"] = [1, 2, {"a": (3, 4)}]
    sh["text"] = "héllo"
    sh["num"] = 3.5
    print(sorted(sh.keys()), len(sh), "text" in sh)
    sh["list"].append(99)
    print(sh["list"])
with shelve.open(path, writeback=True) as sh:
    sh["list"].append(5)
    sh["new"] = {"k": [1]}
with shelve.open(path, "r") as sh:
    print(sh["list"], sh["new"], sh["text"], sorted(sh.items())[0], sh.get("nope", "default"))
    print(sorted(list(sh)), sorted(sh.values(), key=repr)[0])
sh = shelve.open(path)
sh.close()
try:
    sh["x"] = 1
except ValueError as e:
    print(e)
import shutil
shutil.rmtree(d)
