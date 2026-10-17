import datetime
import plistlib

data = {
    "name": "py2go",
    "version": 3,
    "ratio": 0.5,
    "enabled": True,
    "tags": ["a", "b"],
    "blob": b"\x00\x01binary",
    "when": datetime.datetime(2024, 5, 6, 7, 8, 9),
    "nested": {"x": [1, 2, {"y": False}]},
}
xml = plistlib.dumps(data)
print(xml.decode())
back = plistlib.loads(xml)
print(sorted(back.items()) == sorted(data.items()))
print(back["blob"], back["when"], back["nested"])

binary = plistlib.dumps(data, fmt=plistlib.FMT_BINARY)
print(binary[:8], len(binary))
print(plistlib.loads(binary) == data)

print(plistlib.dumps(["x", 1], sort_keys=False).decode().splitlines()[3:7])
try:
    plistlib.loads(b"not a plist")
except plistlib.InvalidFileException as e:
    print("InvalidFileException", e)
uid = plistlib.UID(5)
print(uid, uid.data, plistlib.loads(plistlib.dumps({"u": uid}, fmt=plistlib.FMT_BINARY)))
