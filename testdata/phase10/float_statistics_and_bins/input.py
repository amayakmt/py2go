data = [2.5, 3.7, 1.2, 9.9, 4.4, 4.4, 7.1, 0.3, 6.6, 5.0, 5.0, 8.8]
n = len(data)
mean = sum(data) / n
var = sum((x - mean) ** 2 for x in data) / (n - 1)
sd = var ** 0.5
srt = sorted(data)
median = (srt[n // 2 - 1] + srt[n // 2]) / 2
print(mean, var, sd, median, min(data), max(data), max(data) - min(data), srt[:3], srt[-3:])
bins = {}
for x in data:
    b = int(x // 2.5) * 2.5
    bins[b] = bins.get(b, 0) + 1
for b in sorted(bins):
    print(f"{b:5.1f} - {b + 2.5:5.1f} | {'#' * bins[b]:<6}| {bins[b]:2d} {bins[b] / n:6.1%}")
z = [(x - mean) / sd for x in data]
print([f"{v:+.3f}" for v in z], sum(z) < 1e-12, round(sum(v * v for v in z), 6))
cum = []
acc = 0.0
for x in data:
    acc += x
    cum.append(acc)
print(cum[-1], [round(c, 1) for c in cum], [f"{c:g}" for c in cum[:4]], [f"{c:.3e}" for c in cum[:3]], [f"{c:10.3f}" for c in cum[:2]])
cov = sum((a - mean) * (b - 5.0) for a, b in zip(data, data[::-1])) / (n - 1)
print(cov, cov / (sd * 2.0), round(cov, 9), int(cov * 1000), divmod(cov, 1.5), math_floor := int(cov // 1))
