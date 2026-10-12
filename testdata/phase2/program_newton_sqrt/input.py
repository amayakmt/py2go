def sqrt(x: float) -> float:
    guess: float = x / 2.0
    for _ in range(20):
        guess = (guess + x / guess) / 2.0
    return guess


print(sqrt(2.0))
print(sqrt(144.0))
print(sqrt(0.0001))
