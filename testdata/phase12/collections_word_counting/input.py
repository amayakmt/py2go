from collections import Counter, defaultdict, deque, OrderedDict
text = """the quick brown fox jumps over the lazy dog the fox barks
and the dog sleeps while the quick fox runs"""
words = text.split()
c = Counter(words)
print(c.most_common(3), c["the"], c["cat"], len(c), sum(c.values()), c.total())
by_len = defaultdict(list)
for w in words:
    if w not in by_len[len(w)]:
        by_len[len(w)].append(w)
print(sorted((k, sorted(v)) for k, v in by_len.items()))
first = defaultdict(int)
for i, w in enumerate(words):
    first[w] = first[w] or i + 1
print(sorted(first.items(), key=lambda kv: kv[1])[:4])
letters = Counter(ch for w in words for ch in w)
print(letters.most_common(5), [k for k, v in sorted(letters.items()) if v == 1])
window = deque(maxlen=3)
avgs = []
for n in [1, 5, 3, 8, 2, 9]:
    window.append(n)
    avgs.append(sum(window) / len(window))
print(avgs, list(window))
lru = OrderedDict()
for k in "abcabdcea":
    if k in lru:
        lru.move_to_end(k)
    lru[k] = True
    if len(lru) > 3:
        lru.popitem(last=False)
print(list(lru))
graph = defaultdict(set)
for a, b in [(1, 2), (2, 3), (1, 3), (4, 5)]:
    graph[a].add(b); graph[b].add(a)
seen, queue = {1}, deque([1])
while queue:
    node = queue.popleft()
    for nb in sorted(graph[node]):
        if nb not in seen:
            seen.add(nb); queue.append(nb)
print(sorted(seen), sorted(graph), dict(sorted((k, sorted(v)) for k, v in graph.items())) )
