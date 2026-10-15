import math


def run(steps: int) -> float:
    x, y, vx, vy = 1.0, 0.0, 0.0, 1.0
    for _ in range(steps):
        r = math.sqrt(x * x + y * y)
        ax, ay = -x / (r * r * r), -y / (r * r * r)
        vx += ax * 0.001
        vy += ay * 0.001
        x += vx * 0.001
        y += vy * 0.001
    return x + y


print(round(run(3000000), 6))
