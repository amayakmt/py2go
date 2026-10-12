class Base(Exception):
    pass


class Mid(Base):
    pass


class Leaf(Mid):
    pass


class Other(ValueError):
    pass


def attempt(n: int) -> str:
    try:
        if n == 0:
            raise Leaf("leaf")
        if n == 1:
            raise Mid("mid")
        if n == 2:
            raise Base("base")
        if n == 3:
            raise Other("other")
        if n == 4:
            raise KeyError("k")
        return "ok"
    except Leaf:
        return "Leaf"
    except Mid:
        return "Mid"
    except Base:
        return "Base"
    except ValueError as e:
        return "ValueError:" + str(e)
    except Exception as e:
        return "Exception:" + type(e).__name__


print([attempt(i) for i in range(6)])
