def reverse_words(s: str) -> str:
    words: list[str] = s.split()
    words.reverse()
    return " ".join(words)


print(reverse_words("the sky is blue"))
print(reverse_words("  hello   world  "))
print(repr(reverse_words("")))
