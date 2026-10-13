import string
print(string.ascii_lowercase, string.ascii_uppercase, string.ascii_letters, string.digits, string.hexdigits, string.octdigits)
print(repr(string.punctuation), repr(string.whitespace), len(string.printable), string.capwords("hello   wORLD foo-bar"), string.capwords("a,b c", ","))
print([c for c in "Hello, World 42!" if c in string.ascii_letters], sum(c in string.digits for c in "a1b22"), "".join(sorted(set("banana") & set(string.ascii_lowercase))), string.ascii_letters.index("A"))
