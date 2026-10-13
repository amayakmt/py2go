items = []
total = 0
config = {"debug": False}

def add(x):
    global total
    items.append(x)
    total += x
    return total

def reset():
    global total
    total = 0
    items.clear()
