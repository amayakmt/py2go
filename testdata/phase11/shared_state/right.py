from state import items, add
import state
def count():
    return len(items), state.total
def set_debug():
    state.config["debug"] = True
