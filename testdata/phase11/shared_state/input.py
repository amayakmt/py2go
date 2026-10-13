import left, right, state
left.push(5)
right.add(7)
print(right.count(), state.items, state.total)
right.set_debug()
print(state.config)
state.reset()
print(right.count())
state.total = 100
print(state.total, right.count())
