MODE = "b"
if MODE == "a":
    import plugin_a as plugin
else:
    import plugin_b as plugin
print(plugin.NAME, plugin.run())
for n in range(2):
    from plugin_a import NAME as imported
print(imported)
import plugin_a
print(plugin_a.NAME)
