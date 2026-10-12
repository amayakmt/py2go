class Machine:
    transitions = {
        ("idle", "start"): "running",
        ("running", "pause"): "paused",
        ("paused", "resume"): "running",
        ("running", "stop"): "idle",
        ("paused", "stop"): "idle",
    }

    def __init__(self):
        self.state = "idle"
        self.trace = [self.state]

    def fire(self, event):
        key = (self.state, event)
        if key not in self.transitions:
            raise ValueError("cannot %s while %s" % (event, self.state))
        self.state = self.transitions[key]
        self.trace.append(self.state)
        return self.state

m = Machine()
for ev in ["start", "pause", "start", "resume", "stop", "stop", "start"]:
    try:
        print(ev, "->", m.fire(ev))
    except ValueError as e:
        print("error:", e)
print(m.trace, len(m.trace), m.trace.count("running"))
reachable = {}
for (src, ev), dst in Machine.transitions.items():
    reachable.setdefault(src, []).append(dst)
print({k: sorted(v) for k, v in sorted(reachable.items())})
