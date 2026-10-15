"""getpass: portable password input and user name."""
import os
import sys
import warnings

__all__ = ["getpass", "getuser", "GetPassWarning"]


class GetPassWarning(UserWarning):
    pass


def getpass(prompt="Password: ", stream=None):
    if stream is None:
        stream = sys.stderr
    stream.write(prompt)
    stream.flush()
    line = sys.stdin.readline()
    if not line:
        raise EOFError
    if line[-1] == "\n":
        line = line[:-1]
    return line


def getuser():
    for name in ("LOGNAME", "USER", "LNAME", "USERNAME"):
        user = os.environ.get(name)
        if user:
            return user
    import pwd
    return pwd.getpwuid(os.getuid())[0]
