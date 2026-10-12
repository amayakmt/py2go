class InsufficientFunds(Exception):
    def __init__(self, needed, available):
        super().__init__("need %.2f, have %.2f" % (needed, available))
        self.needed = needed
        self.available = available

class Account:
    count = 0

    def __init__(self, owner, balance=0):
        Account.count += 1
        self.id = Account.count
        self.owner = owner
        self.balance = balance
        self.history = []

    def deposit(self, amount):
        if amount <= 0:
            raise ValueError("deposit must be positive")
        self.balance += amount
        self.history.append(("deposit", amount))
        return self.balance

    def withdraw(self, amount):
        if amount > self.balance:
            raise InsufficientFunds(amount, self.balance)
        self.balance -= amount
        self.history.append(("withdraw", amount))
        return self.balance

    def transfer(self, other, amount):
        self.withdraw(amount)
        other.deposit(amount)

    def __str__(self):
        return "#%d %s: %.2f" % (self.id, self.owner, self.balance)

a, b = Account("ann", 100), Account("bob")
a.deposit(50)
a.transfer(b, 30.5)
print(a, b, Account.count)
for action in (lambda: a.withdraw(500), lambda: b.deposit(-5), lambda: a.transfer(b, 119.5), lambda: a.transfer(b, 0.01)):
    try:
        action()
        print("ok")
    except InsufficientFunds as e:
        print("insufficient:", e, e.needed, e.available)
    except ValueError as e:
        print("invalid:", e)
print(a, b)
print(a.history, b.history)
print(sum(x.balance for x in (a, b)), max((a, b), key=lambda x: x.balance).owner)
