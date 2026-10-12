def withdraw(balance: int, amount: int) -> int:
    if amount > balance:
        raise ValueError(f"insufficient funds: {balance} < {amount}")
    return balance - amount


for amount in [30, 80, 70]:
    try:
        print(withdraw(50, amount))
    except ValueError as e:
        print(f"error: {e}")
