try:
    raise ValueError("bad") from None
except (ValueError, TypeError) as e:
    raise
except:
    pass
else:
    assert 1 < 2, "never"
finally:
    print("done")
try:
    pass
finally:
    pass
