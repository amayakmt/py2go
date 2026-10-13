from .. import NAME
from ..util import helpers
from ..util.helpers import shout as s2
class Engine:
    def run(self, x):
        return helpers.shout(f"{NAME}:{x}") + s2("!")
