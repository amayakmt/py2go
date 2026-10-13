import random
random.seed(2024)
xs = list(range(20))
random.shuffle(xs)
print(xs)
ys = list("abcdefgh"); random.shuffle(ys); print(ys)
empty = []; random.shuffle(empty); one = [1]; random.shuffle(one); print(empty, one)
print(random.sample(range(1000), 5), random.sample(range(10), 10), random.sample("hello world", 4), random.sample([1, 2, 3], 0), random.sample(range(10 ** 6), 8), random.sample(list(range(100)), 30))
print(random.sample(["a", "b", "c"], counts=[3, 0, 2], k=4), random.sample(range(5), counts=[1, 1, 1, 1, 6], k=3))
pop = ["red", "green", "blue"]
print(random.choices(pop, k=8), random.choices(pop, weights=[10, 1, 1], k=8), random.choices(pop, cum_weights=[1, 2, 10], k=8), random.choices(range(5)), random.choices(pop, k=0))
for bad in (lambda: random.choices(pop, weights=[1, 2]), lambda: random.choices(pop, weights=[0, 0, 0]), lambda: random.choices(pop, weights=[1, 1, 1], cum_weights=[1, 2, 3]), lambda: random.choices([], k=2), lambda: random.shuffle("abc")):
    try:
        print(bad())
    except (ValueError, TypeError, IndexError) as e:
        print(type(e).__name__, e)
deck = [(r, s) for r in "A23456789TJQK" for s in "SHDC"]
random.shuffle(deck)
hand = deck[:5]
print(hand, len(set(deck)), sorted(hand)[:2])
counts = {}
for _ in range(1000):
    k = random.randint(1, 4)
    counts[k] = counts.get(k, 0) + 1
print(sorted(counts.items()), sum(counts.values()))
print(random.triangular(0, 1), random.triangular(5, 10, 6), random.triangular(2, 2, 2), random.paretovariate(2.0) > 0, random.weibullvariate(1, 1) > 0)
