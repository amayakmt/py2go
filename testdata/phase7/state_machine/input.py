class State:
    name: str = "state"

    def next(self, event: str) -> "State":
        return self


class Idle(State):
    name = "idle"

    def next(self, event: str) -> State:
        if event == "start":
            return Running()
        return self


class Running(State):
    name = "running"

    def next(self, event: str) -> State:
        if event == "pause":
            return Paused()
        if event == "stop":
            return Idle()
        return self


class Paused(State):
    name = "paused"

    def next(self, event: str) -> State:
        if event == "resume":
            return Running()
        if event == "stop":
            return Idle()
        return self


s: State = Idle()
for ev in ["start", "pause", "bogus", "resume", "stop", "stop"]:
    s = s.next(ev)
    print(ev, "->", s.name, type(s).__name__)
