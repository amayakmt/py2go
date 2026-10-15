import argparse

p = argparse.ArgumentParser(prog="tool")
p.add_argument("path")
p.add_argument("--n", type=int, default=2)
print(p.parse_args(["a", "--n", "5"]))
print(p.parse_args(["b"]))
p.parse_args(["--n", "3"])
print("unreachable")
