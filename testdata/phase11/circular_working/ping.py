import pong
COUNT = 0
def ping(n):
    global COUNT
    COUNT += 1
    if n <= 0:
        return "done"
    return pong.pong(n - 1)
