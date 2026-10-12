def digit_sum(n: int) -> int:
    total: int = 0
    while n > 0:
        total += n % 10
        n //= 10
    return total


print(digit_sum(12345))
print(digit_sum(999999999))
print(digit_sum(0))
