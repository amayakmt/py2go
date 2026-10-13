print("start")
def use():
    import lazy
    return lazy.work()
def use2():
    from lazy import work
    return work()
print(use(), use2())
