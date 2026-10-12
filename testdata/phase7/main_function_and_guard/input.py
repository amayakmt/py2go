class App:
    def __init__(self, name: str) -> None:
        self.name = name

    def main(self) -> str:
        return "app " + self.name


def main() -> None:
    app: App = App("demo")
    print(app.main())
    print(helper(3))


def helper(n: int) -> int:
    return n * 2


if __name__ == "__main__":
    main()
    print("done")
