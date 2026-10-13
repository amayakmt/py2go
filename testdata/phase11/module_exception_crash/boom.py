class Boom(Exception):
    pass
def go():
    raise Boom("it broke")
