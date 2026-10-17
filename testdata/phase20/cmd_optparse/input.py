import cmd
import io
from optparse import OptionParser, OptionGroup, OptionValueError


class Shell(cmd.Cmd):
    prompt = "(sh) "
    intro = "welcome"

    def do_greet(self, arg):
        """greet [name]: say hello"""
        print("hello", arg or "world")

    def do_add(self, arg):
        print(sum(int(x) for x in arg.split()))

    def complete_greet(self, text, line, begidx, endidx):
        return [n for n in ("alice", "bob") if n.startswith(text)]

    def do_quit(self, arg):
        return True

    def default(self, line):
        print("unknown:", line)

    def emptyline(self):
        print("(empty)")


out = io.StringIO()
sh = Shell(stdin=io.StringIO("greet\ngreet bob\nadd 1 2 3\n\nfoo bar\nhelp greet\n?\nquit\n"), stdout=out)
sh.use_rawinput = False
sh.cmdloop()
print(out.getvalue())
print(sh.onecmd("greet carol"))
print(sh.completenames("gr"), sh.complete_greet("a", "greet a", 6, 7))
print(sh.parseline("add 4 5"))

parser = OptionParser(usage="%prog [options] file", version="%prog 1.0", prog="tool")
parser.add_option("-v", "--verbose", action="store_true", dest="verbose", default=False, help="talk more")
parser.add_option("-n", "--num", type="int", dest="num", default=1, help="how many [default: %default]")
parser.add_option("-o", dest="out", metavar="FILE")
parser.add_option("--tag", action="append", dest="tags")
parser.add_option("-c", action="count", dest="count")
group = OptionGroup(parser, "Extra", "rarely used")
group.add_option("--level", type="choice", choices=["low", "high"], default="low")
parser.add_option_group(group)
opts, args = parser.parse_args(["-v", "-n", "3", "-ofile.txt", "--tag", "a", "--tag=b", "-cc", "--level", "high", "x", "y"])
print(opts.verbose, opts.num, opts.out, opts.tags, opts.count, opts.level, args)
print(parser.format_help())
for bad in (["-n", "x"], ["--level", "mid"], ["--nope"]):
    try:
        parser.parse_args(bad)
    except SystemExit as e:
        print("exit", e.code)


def check(option, opt_str, value, parser):
    if value < 0:
        raise OptionValueError("%s must be >= 0" % opt_str)
    parser.values.size = value


p2 = OptionParser()
p2.add_option("--size", type="int", action="callback", callback=check)
print(p2.parse_args(["--size", "5"])[0].size)
