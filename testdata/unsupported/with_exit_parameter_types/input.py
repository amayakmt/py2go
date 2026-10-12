class C:
    def __enter__(self) -> int:
        return 1

    def __exit__(self, t: int, e: int, tb: int) -> None:
        pass


with C():
    pass
