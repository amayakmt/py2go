from typing import Callable


def make_account(balance: int) -> tuple[Callable[[int], int], Callable[[], int]]:
    def deposit(n: int) -> int:
        nonlocal balance
        balance += n
        return balance

    def read() -> int:
        return balance

    return (deposit, read)


deposit, read = make_account(100)
print(deposit(50), read(), deposit(-30), read())
