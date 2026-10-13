import statistics as st
from fractions import Fraction

data = [2.5, 3.1, 4.7, 5.0, 5.0, 8.2, 9.9]
ints = [1, 2, 3, 4, 5, 6]
print(st.mean(data), st.mean(ints), st.mean([1, 2, 3]), st.fmean(data), st.fmean(ints), st.geometric_mean([1, 2, 4, 8]), st.harmonic_mean([1, 2, 4]))
print(st.median(data), st.median(ints), st.median_low(ints), st.median_high(ints), st.median_grouped([1, 2, 2, 3, 4, 7, 9]), st.median_grouped([52, 52, 53, 54], 2))
print(st.mode(["a", "b", "a", "c"]), st.mode(data), st.multimode("aabbcc"), st.multimode([]))
print(st.variance(data), st.pvariance(data), st.stdev(data), st.pstdev(data), st.variance(ints), st.stdev(ints), st.pstdev(ints))
print(st.variance([Fraction(1, 3), Fraction(1, 2), Fraction(5, 6)]), st.mean([Fraction(1, 3), Fraction(1, 2)]), st.mean([1, 2.5, Fraction(1, 2)]))
print(st.quantiles(data), st.quantiles(data, n=10), st.quantiles(data, method="inclusive"), st.quantiles([1, 2, 3, 4, 5], n=2))
x = [1, 2, 3, 4, 5, 6, 7, 8, 9]
y = [1, 2, 1, 3, 3, 1, 4, 5, 6]
print(st.correlation(x, y), st.covariance(x, y), st.correlation(x, y, method="ranked"))
r = st.linear_regression(x, y)
print(r.slope, r.intercept, r, st.linear_regression(x, y, proportional=True))
print(st.mean([1e308, 1e308]) if False else st.mean([1e15, 1.0, -1e15]), st.mean([0.1] * 10), st.fmean([0.1] * 10), st.pvariance([1e9 + 1, 1e9 + 2, 1e9 + 3]))
for f, args in [(st.mean, ([],)), (st.median, ([],)), (st.mode, ([],)), (st.variance, ([1],)), (st.pvariance, ([],)), (st.stdev, ([1],)), (st.geometric_mean, ([-1, 1],)), (st.geometric_mean, ([],)), (st.geometric_mean, ([0, 1],)), (st.harmonic_mean, ([-1, 2],)),
                (st.quantiles, ([1],)), (st.correlation, ([1, 1], [1, 1])), (st.fmean, ([],)), (st.mean, (["a"],))]:
    try:
        print(f(*args))
    except (st.StatisticsError, TypeError, ValueError) as e:
        print(type(e).__name__, e)
print(issubclass(st.StatisticsError, ValueError))
nd = st.NormalDist(100, 15)
print(nd, nd.mean, nd.stdev, nd.variance, nd.pdf(100), nd.cdf(115), nd.zscore(130), nd + 10, nd * 2, nd - st.NormalDist(1, 1), nd == st.NormalDist(100.0, 15.0))
print(st.NormalDist.from_samples([1, 2, 3, 4]))
