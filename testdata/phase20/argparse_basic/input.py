import argparse

p = argparse.ArgumentParser(prog="tool", description="A tool.")
p.add_argument("path", help="input path")
p.add_argument("-n", "--count", type=int, default=1, help="how many")
p.add_argument("-v", "--verbose", action="store_true")
p.add_argument("--mode", choices=["fast", "slow"], default="fast")
p.add_argument("-x", action="append", default=[])
p.add_argument("--level", action="count", default=0)
p.add_argument("rest", nargs="*")

ns = p.parse_args(["in.txt", "-n", "3", "-v", "--mode", "slow", "-x", "a", "-x", "b", "--level", "--level", "r1", "r2"])
print(ns)
print(ns.path, ns.count, ns.verbose, ns.mode, ns.x, ns.level, ns.rest)
print(vars(p.parse_args(["f"])))
print(p.parse_known_args(["f", "--unknown", "1"]))
print(p.format_usage())
print(p.format_help())

sub = argparse.ArgumentParser(prog="git")
subs = sub.add_subparsers(dest="cmd")
a = subs.add_parser("add", help="add files")
a.add_argument("files", nargs="+")
c = subs.add_parser("commit")
c.add_argument("-m", "--message", required=True)
print(sub.parse_args(["add", "a", "b"]))
print(sub.parse_args(["commit", "-m", "msg"]))

g = argparse.ArgumentParser(prog="g")
grp = g.add_mutually_exclusive_group()
grp.add_argument("--a", action="store_true")
grp.add_argument("--b", action="store_true")
g.add_argument("--n", type=float, default=0.5)
g.add_argument("--flag", action=argparse.BooleanOptionalAction, default=False)
print(g.parse_args(["--a", "--n", "2.5", "--flag"]))
print(g.parse_args(["--no-flag"]))

def attempt(argv):
    try:
        g.parse_args(argv)
    except SystemExit as e:
        print("exit", e.code)


attempt(["--a", "--b"])
attempt(["--n", "x"])
attempt(["--bogus"])

q = argparse.ArgumentParser(prog="q", exit_on_error=False)
q.add_argument("--k", type=int, required=True)
try:
    q.parse_args([])
except argparse.ArgumentError as e:
    print("ArgumentError", e)
except SystemExit as e:
    print("exit", e.code)
print(argparse.Namespace(a=1, b="x"))
