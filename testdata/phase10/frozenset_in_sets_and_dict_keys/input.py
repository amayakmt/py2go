edges = [frozenset((1, 2)), frozenset((2, 1)), frozenset((2, 3)), frozenset((3,)), frozenset((3, 3))]
print(len(set(edges)), sorted(map(sorted, set(edges))), edges[0] == edges[1], edges[3] == edges[4])
graph = {}
for e in edges:
    graph[e] = graph.get(e, 0) + 1
print(sorted((sorted(k), v) for k, v in graph.items()))
groups = {}
for word in ["listen", "silent", "enlist", "google", "gogole", "cat", "act", "tac", "dog"]:
    groups.setdefault(frozenset(word), []).append(word)
print(sorted(sorted(v) for v in groups.values()), len(groups))
powerset = [frozenset()]
for x in [1, 2, 3]:
    powerset += [s | {x} for s in powerset]
print(len(powerset), sorted(sorted(s) for s in powerset), frozenset(powerset) == frozenset(powerset[::-1]))
nested = frozenset([frozenset([1, 2]), frozenset([3]), frozenset()])
print(len(nested), frozenset([1, 2]) in nested, frozenset() in nested, sorted(len(x) for x in nested), sum(len(x) for x in nested))
visited = set()
visited.add(frozenset({1, 2}))
print(frozenset({2, 1}) in visited, frozenset({1, 2}) == {1, 2})
