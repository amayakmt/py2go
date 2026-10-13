from .text import count_words
def stats(xs):
    return min(xs), max(xs), sum(xs) / len(xs), count_words("a b c")
