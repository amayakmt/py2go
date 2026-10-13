import shapes
from shapes import Circle, square
from shapes.square import side_area
import shapes.circle as sc
print(shapes.DEFAULT, Circle(2).area(), square.side_area(3), side_area(4))
print(shapes.square.as_circle(4).area(), shapes.circle.Circle(1).r)
from shapes import *
print(DEFAULT, Circle(1).area())
