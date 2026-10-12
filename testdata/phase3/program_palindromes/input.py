def is_palindrome(s: str) -> bool:
    cleaned: str = ""
    for c in s.lower():
        if c in "abcdefghijklmnopqrstuvwxyz0123456789":
            cleaned += c
    return cleaned == cleaned[::-1]


for s in ["A man, a plan, a canal: Panama", "hello", "Was it a car or a cat I saw?", ""]:
    print(f"{s!r}: {is_palindrome(s)}")
