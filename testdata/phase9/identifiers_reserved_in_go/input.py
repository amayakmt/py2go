default = 1
var = 2
func = lambda x: x + var
chan = [default, var]
interface = {"k": chan}
package = "pkg"
select = len(chan)
struct = (default, var)
switch = default if var else 0
string = "text"
print(default, var, func(5), chan, interface, package, select, struct, switch, string)

def go(defer, fallthrough=3, *goto, **const):
    return defer + fallthrough + len(goto) + len(const)

print(go(1), go(1, 2, 3, 4, a=1), go(defer=5, fallthrough=0))

class Holder:
    def __init__(self, range_, type_):
        self.range = range_
        self.type = type_

    def map(self, f):
        return [f(x) for x in self.range]

    def default(self):
        return self.type

h = Holder([1, 2, 3], "T")
print(h.map(lambda x: x * 2), h.default(), h.range, h.type)
def switch_case(case, int64=0):
    nil = None
    true = False
    return (case, int64, nil, true)
print(switch_case(1), switch_case(2, int64=5))
