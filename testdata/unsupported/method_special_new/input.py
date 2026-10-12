"""x"""
class A:
    def __new__(cls) -> "A":
        return object.__new__(cls)
