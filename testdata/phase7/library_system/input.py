class BookNotAvailable(Exception):
    pass


class MemberLimit(Exception):
    def __init__(self, member: str, limit: int) -> None:
        super().__init__(f"{member} reached the limit of {limit}")
        self.member = member


class Book:
    def __init__(self, isbn: str, title: str, copies: int = 1) -> None:
        self.isbn = isbn
        self.title = title
        self.copies = copies

    def __str__(self) -> str:
        return f"{self.title} [{self.isbn}] x{self.copies}"


class Member:
    LIMIT: int = 2

    def __init__(self, name: str) -> None:
        self.name = name
        self.borrowed: list[Book] = []

    def borrow(self, book: Book) -> None:
        if len(self.borrowed) >= Member.LIMIT:
            raise MemberLimit(self.name, Member.LIMIT)
        if book.copies == 0:
            raise BookNotAvailable(book.title)
        book.copies -= 1
        self.borrowed.append(book)

    def give_back(self, isbn: str) -> bool:
        for b in self.borrowed:
            if b.isbn == isbn:
                b.copies += 1
                self.borrowed.remove(b)
                return True
        return False


class Library:
    def __init__(self) -> None:
        self.books: dict[str, Book] = {}
        self.members: dict[str, Member] = {}

    def add_book(self, b: Book) -> None:
        self.books[b.isbn] = b

    def join(self, name: str) -> Member:
        m: Member = Member(name)
        self.members[name] = m
        return m

    def report(self) -> list[str]:
        return [str(b) for b in self.books.values()]


lib: Library = Library()
lib.add_book(Book("1", "Dune"))
lib.add_book(Book("2", "Emma", 2))
lib.add_book(Book("3", "Ulysses"))
ann: Member = lib.join("ann")
bob: Member = lib.join("bob")
attempts: list[tuple[Member, str]] = [(ann, "1"), (bob, "1"), (ann, "2"), (ann, "3"), (bob, "2")]
for who, isbn in attempts:
    try:
        who.borrow(lib.books[isbn])
        print(who.name, "borrowed", lib.books[isbn].title)
    except BookNotAvailable as e:
        print("not available:", e)
    except MemberLimit as e:
        print("limit:", e, e.member)
print(lib.report())
print(ann.give_back("1"), ann.give_back("9"), lib.report())
