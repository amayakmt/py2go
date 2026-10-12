class Account:
    def __init__(self, owner: str, balance: int) -> None:
        self.owner = owner
        self.balance = balance
        self.history: list[str] = []

    def deposit(self, amount: int) -> None:
        self.balance += amount
        self.log("deposit " + str(amount))

    def withdraw(self, amount: int) -> bool:
        if amount > self.balance:
            self.log("denied " + str(amount))
            return False
        self.balance -= amount
        self.log("withdraw " + str(amount))
        return True

    def log(self, text: str) -> None:
        self.history.append(text)

    def summary(self) -> str:
        return f"{self.owner}: {self.balance} ({len(self.history)} events)"


acct: Account = Account("ann", 100)
acct.deposit(50)
print(acct.withdraw(500), acct.withdraw(30))
print(acct.summary())
print(acct.history)
