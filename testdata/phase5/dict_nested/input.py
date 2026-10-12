users: dict[str, dict[str, int]] = {
    "ann": {"age": 30, "score": 90},
    "bob": {"age": 25, "score": 85},
}
users["ann"]["score"] += 5
users["cy"] = {"age": 41}
users["cy"]["score"] = 70
for name in users:
    print(name, users[name], users[name]["age"])
graph: dict[int, list[int]] = {1: [2, 3], 2: [4], 3: [4], 4: []}
for node in graph:
    for nxt in graph[node]:
        print(node, "->", nxt)
print(users)
