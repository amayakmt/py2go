def cmd_add(args):
    return sum(int(a) for a in args)

def cmd_echo(args):
    return " ".join(args)

def cmd_rev(args):
    return " ".join(reversed(args))

def cmd_help(args):
    return "commands: " + ", ".join(sorted(COMMANDS))

COMMANDS = {"add": cmd_add, "echo": cmd_echo, "rev": cmd_rev, "help": cmd_help}

def dispatch(line):
    parts = line.split()
    if not parts:
        return None
    name, args = parts[0], parts[1:]
    handler = COMMANDS.get(name)
    if handler is None:
        return "unknown command: " + name
    try:
        return handler(args)
    except ValueError as e:
        return "bad arguments: " + str(e)

for line in ("add 1 2 3", "echo hi there", "rev a b c", "help", "", "nope", "add 1 x"):
    print(repr(line), "->", repr(dispatch(line)))
COMMANDS["twice"] = lambda args: [a * 2 for a in args]
print(dispatch("twice ab c"), len(COMMANDS))
