import first
print("second sees early:", first.early)
try:
    print(first.late)
except AttributeError as e:
    print("AttributeError")
try:
    from first import late
except ImportError as e:
    print("ImportError")
