import cfg
def bump():
    cfg.level += 1
    cfg.name = cfg.name + "!"
    return cfg.level
print(cfg.show(), bump(), bump(), cfg.show())
cfg.level = 50
print(cfg.show())
