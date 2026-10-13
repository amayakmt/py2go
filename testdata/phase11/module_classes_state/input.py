import registry
import shapes2
print(registry.names(), registry.create("sq", 3).area(), registry.create("rect", 2, 5).area())
print(type(registry.create("sq", 1)).__name__)
