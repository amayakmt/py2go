import errors
from errors import NotFound, find
for v in (1, 0, -3):
    try:
        print(find(v))
    except NotFound as e:
        print("NotFound", e, e.what, type(e).__name__, isinstance(e, errors.AppError))
    except errors.AppError as e:
        print("AppError", e, repr(e))
print(issubclass(NotFound, errors.AppError), NotFound.__name__)
find(-1)
