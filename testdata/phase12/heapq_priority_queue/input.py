import heapq
h = []
for x in [9, 4, 7, 1, 8, 2, 6, 3, 5, 0]:
    heapq.heappush(h, x)
print(h, h[0], len(h))
print([heapq.heappop(h) for _ in range(4)], h)
data = [5, 3, 8, 1, 9, 2, 7]
heapq.heapify(data)
print(data, heapq.heappushpop(data, 0), heapq.heappushpop(data, 4), heapq.heapreplace(data, 10), data)
print(heapq.nlargest(3, [4, 9, 1, 7, 6]), heapq.nsmallest(3, [4, 9, 1, 7, 6]), heapq.nlargest(2, ["pear", "fig", "banana"], key=len), heapq.nsmallest(2, [(2, "b"), (1, "z"), (1, "a")]))
print(list(heapq.merge([1, 4, 9], [2, 3, 10], [0])), list(heapq.merge([5, 3, 1], [4, 2], reverse=True)))
tasks = []
for prio, name in [(3, "low"), (1, "urgent"), (2, "normal"), (1, "also urgent")]:
    heapq.heappush(tasks, (prio, name))
print([heapq.heappop(tasks)[1] for _ in range(4)])
try:
    heapq.heappop([])
except IndexError as e:
    print(e)
words = "the quick brown fox jumps over the lazy dog".split()
print(heapq.nsmallest(3, words), heapq.nlargest(3, words, key=lambda w: (len(w), w)))
