import json, os
path = "/tmp/py2go_p12_json.json"
records = [{"id": i, "name": "item%d" % i, "price": i * 1.25, "tags": ["t%d" % j for j in range(i % 3)], "meta": None if i % 2 else {"ok": True}} for i in range(5)]
with open(path, "w") as f:
    json.dump(records, f, indent=2)
with open(path) as f:
    text = f.read()
print(len(text), text[:60].replace("\n", "|"))
with open(path) as f:
    loaded = json.load(f)
print(loaded == records, [r["name"] for r in loaded], sum(r["price"] for r in loaded), loaded[3]["meta"], loaded[4]["meta"])
by_id = {r["id"]: r for r in loaded}
by_id[2]["tags"].append("extra")
print(json.dumps(by_id[2], sort_keys=True), json.dumps(list(by_id)), json.loads(json.dumps(by_id)).keys() if False else sorted(json.loads(json.dumps(by_id))))
os.remove(path)
config = json.loads('{"debug": true, "ports": [80, 443], "paths": {"root": "/", "tmp": "/tmp"}}')
config["ports"].append(8080)
print(json.dumps(config, sort_keys=True), config["paths"]["tmp"], config.get("missing", "dflt"), type(config["debug"]).__name__)
print(json.dumps({"unicode": "é", "emoji": "😀", "quote": "\"", "nul": "\x00"}), json.loads(json.dumps("\ud7ff\ue000")) == "\ud7ff\ue000")
