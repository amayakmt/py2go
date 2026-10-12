class InsufficientFunds(Exception):
    def __init__(self, needed: int, available: int) -> None:
        super().__init__(f"need {needed}, have {available}")
        self.needed = needed
        self.available = available


class Account:
    next_id: int = 1

    def __init__(self, owner: str, balance: int = 0) -> None:
        self.id = Account.next_id
        Account.next_id += 1
        self.owner = owner
        self.balance = balance

    def withdraw(self, amount: int) -> None:
        if amount > self.balance:
            raise InsufficientFunds(amount, self.balance)
        self.balance -= amount

    def deposit(self, amount: int) -> None:
        self.balance += amount

    def __repr__(self) -> str:
        return f"Account#{self.id}({self.owner}, {self.balance})"


class Bank:
    def __init__(self) -> None:
        self.accounts: dict[int, Account] = {}

    def open(self, owner: str, initial: int) -> Account:
        acct: Account = Account(owner, initial)
        self.accounts[acct.id] = acct
        return acct

    def transfer(self, src: int, dst: int, amount: int) -> bool:
        try:
            self.accounts[src].withdraw(amount)
        except InsufficientFunds as e:
            print("transfer failed:", e, e.needed - e.available)
            return False
        self.accounts[dst].deposit(amount)
        return True


bank: Bank = Bank()
a: Account = bank.open("ann", 100)
b: Account = bank.open("bob", 20)
print(bank.transfer(a.id, b.id, 70), bank.transfer(a.id, b.id, 70))
print(a, b, bank.accounts)
