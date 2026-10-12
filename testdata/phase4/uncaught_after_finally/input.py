def run() -> None:
    try:
        print("working")
        raise RuntimeError("fatal")
    finally:
        print("cleanup")


print("start")
run()
print("never")
