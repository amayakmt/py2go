def run(program, inputs=()):
    stack = []
    out = []
    pc = 0
    inputs = list(inputs)
    steps = 0
    while pc < len(program):
        op, *args = program[pc]
        steps += 1
        if steps > 1000:
            raise RuntimeError("too many steps")
        if op == "push":
            stack.append(args[0])
        elif op == "pop":
            stack.pop()
        elif op == "add":
            b, a = stack.pop(), stack.pop()
            stack.append(a + b)
        elif op == "mul":
            b, a = stack.pop(), stack.pop()
            stack.append(a * b)
        elif op == "dup":
            stack.append(stack[-1])
        elif op == "read":
            stack.append(inputs.pop(0))
        elif op == "print":
            out.append(stack.pop())
        elif op == "jnz":
            if stack.pop() != 0:
                pc = args[0]
                continue
        elif op == "dec":
            stack.append(stack.pop() - 1)
        else:
            raise ValueError("bad op " + op)
        pc += 1
    return out, stack

print(run([("push", 2), ("push", 3), ("add",), ("push", 4), ("mul",), ("print",)]))
print(run([("read",), ("read",), ("add",), ("print",)], [10, 32]))
countdown = [("push", 3), ("dup",), ("print",), ("dec",), ("dup",), ("jnz", 1)]
print(run(countdown))
for bad in ([("nope",)], [("add",)], [("push", 1), ("jnz", 0)]):
    try:
        print(run(bad))
    except (ValueError, IndexError, RuntimeError) as e:
        print(type(e).__name__, e)
