import scopes
from scopes import f, g, h, K
x = "main x"
print(scopes.x, x, f(1), g(), h(), scopes.comp(), scopes.lam(), K.z, K.y, K().m(), scopes.walrus(), scopes.names)
print(scopes.caught, scopes.last, scopes.i)
