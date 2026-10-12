graph = {
    "a": ["b", "c"],
    "b": ["d"],
    "c": ["d", "e"],
    "d": ["f"],
    "e": ["f"],
    "f": [],
}

def bfs(start, goal):
    queue = [[start]]
    seen = {start}
    while queue:
        path = queue.pop(0)
        node = path[-1]
        if node == goal:
            return path
        for nxt in graph[node]:
            if nxt not in seen:
                seen.add(nxt)
                queue.append(path + [nxt])
    return None

def dfs(node, seen=None):
    seen = seen if seen is not None else []
    seen.append(node)
    for nxt in graph[node]:
        if nxt not in seen:
            dfs(nxt, seen)
    return seen

def topo():
    indeg = {n: 0 for n in graph}
    for n in graph:
        for m in graph[n]:
            indeg[m] += 1
    ready = sorted(n for n, d in indeg.items() if d == 0)
    order = []
    while ready:
        n = ready.pop(0)
        order.append(n)
        for m in graph[n]:
            indeg[m] -= 1
            if indeg[m] == 0:
                ready.append(m)
                ready.sort()
    return order

print(bfs("a", "f"), bfs("a", "a"), bfs("f", "a"))
print(dfs("a"), topo())
print(sum(len(v) for v in graph.values()), [n for n in graph if not graph[n]])
