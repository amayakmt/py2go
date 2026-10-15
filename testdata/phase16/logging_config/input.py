import logging
import logging.config
import sys

logging.config.dictConfig({
    "version": 1,
    "disable_existing_loggers": False,
    "formatters": {
        "plain": {"format": "%(levelname)s|%(name)s|%(message)s"},
        "braces": {"format": "{levelname}: {message}", "style": "{"},
    },
    "filters": {"onlyapp": {"name": "app"}},
    "handlers": {
        "out": {"class": "logging.StreamHandler", "formatter": "plain", "stream": "ext://sys.stdout", "level": "INFO"},
        "alt": {"class": "logging.StreamHandler", "formatter": "braces", "stream": "ext://sys.stdout", "filters": ["onlyapp"]},
        "nul": {"class": "logging.NullHandler"},
    },
    "loggers": {
        "app": {"level": "DEBUG", "handlers": ["out", "alt"], "propagate": False},
        "app.sub": {"level": "WARNING"},
        "quiet": {"handlers": ["nul"], "propagate": False},
    },
    "root": {"level": "ERROR", "handlers": ["out"]},
})
app = logging.getLogger("app")
app.debug("hidden by out level but not alt")
app.info("info")
logging.getLogger("app.sub").info("filtered by app.sub level")
logging.getLogger("app.sub").warning("sub warning")
logging.getLogger("other").error("other error")
logging.getLogger("other").warning("other warning")
logging.getLogger("quiet").error("quiet")
print(logging.getLogger("app").handlers, logging.getLogger().level, logging.getLogger("app").propagate)
for bad in [{}, {"version": 2}, {"version": 1, "handlers": {"h": {"class": "nope.Handler"}}},
            {"version": 1, "handlers": {"h": {"class": "logging.StreamHandler", "formatter": "missing"}}}]:
    try:
        logging.config.dictConfig(bad)
    except ValueError as e:
        print(e)
