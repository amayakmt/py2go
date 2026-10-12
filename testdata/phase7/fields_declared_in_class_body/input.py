class Particle:
    x: float
    y: float
    mass: float = 1.0
    tags: list[str]

    def __init__(self, x: float, y: float) -> None:
        self.x = x
        self.y = y
        self.tags = []

    def energy(self, v: float) -> float:
        return 0.5 * self.mass * v * v


p: Particle = Particle(1.0, 2.0)
p.tags.append("fast")
p.mass = 4.0
print(p.x, p.y, p.mass, p.tags, p.energy(3.0))
q: Particle = Particle(0.0, 0.0)
print(q.mass, Particle.mass)
