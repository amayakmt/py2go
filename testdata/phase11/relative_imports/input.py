from app.core import Engine, shout
import app.core.engine
from app import NAME
print(Engine().run("go"), shout("x"), NAME)
import app.util.helpers as h
print(h.shout("done"), h.__name__, app.core.engine.__name__, app.__name__)
