import sys
print(sys.version_info >= (3, 8), sys.version_info[:2], sys.version_info.major, sys.version_info.minor, sys.platform, sys.maxsize, sys.maxsize == 2 ** 63 - 1, sys.byteorder, sys.maxunicode)
print(sys.getrecursionlimit(), len(sys.argv) >= 1, sys.float_info.max, sys.float_info.epsilon, sys.float_info.dig, sys.int_info.default_max_str_digits, sys.hexversion >= 0x03080000)
sys.setrecursionlimit(5000)
print(sys.getrecursionlimit(), sys.getdefaultencoding(), sys.stdout is sys.__stdout__)
sys.stdout.write("direct write\n")
n = sys.stdout.write("count\n")
print(n)
print("to stdout", file=sys.stdout)
print("a", "b", sep="-", end="!\n", file=sys.stdout)
print("to stderr", file=sys.stderr)
sys.stderr.write("err write\n")
print(sys.stdout.name, sys.stdout.closed, sys.stdout.writable(), sys.stdin.readable())
try:
    sys.setrecursionlimit(0)
except ValueError as e:
    print(e)
def main():
    print("exiting")
    sys.exit(0)
try:
    main()
except SystemExit as e:
    print("SystemExit", e.code if hasattr(e, "code") else e.args)
try:
    sys.exit("bye")
except SystemExit as e:
    print(e.args)
