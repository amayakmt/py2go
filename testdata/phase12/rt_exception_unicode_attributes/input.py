try:
    b"\xff".decode()
except UnicodeDecodeError as e:
    print(e.start)
