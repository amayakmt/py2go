class Queue:
    def __init__(self):
        self.items = []

    def push(self, x):
        self.items.append(x)

    def pop(self):
        if not self.items:
            raise IndexError("pop from empty queue")
        return self.items.pop(0)

    def __len__(self):
        return len(self.items)

    def __bool__(self):
        return bool(self.items)

    def __iter__(self):
        return iter(list(self.items))

class Customer:
    def __init__(self, name, arrival, service):
        self.name = name
        self.arrival = arrival
        self.service = service

    def __repr__(self):
        return "%s@%d" % (self.name, self.arrival)

customers = [Customer("a", 0, 3), Customer("b", 1, 2), Customer("c", 2, 1), Customer("d", 10, 2)]
waiting = Queue()
clock = 0
done = []
pending = list(customers)
while pending or waiting:
    while pending and pending[0].arrival <= clock:
        waiting.push(pending.pop(0))
    if waiting:
        c = waiting.pop()
        start = clock
        clock += c.service
        done.append((c.name, start, clock, start - c.arrival))
    else:
        clock = pending[0].arrival
print(done)
waits = [d[3] for d in done]
print(sum(waits) / len(waits), max(waits), clock, list(waiting), len(waiting))
q = Queue()
try:
    q.pop()
except IndexError as e:
    print(e)
