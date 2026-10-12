def caesar(text, shift):
    out = []
    for ch in text:
        if ch.isalpha():
            base = ord("A") if ch.isupper() else ord("a")
            out.append(chr((ord(ch) - base + shift) % 26 + base))
        else:
            out.append(ch)
    return "".join(out)

def vigenere(text, key, decrypt=False):
    out = []
    k = 0
    for ch in text:
        if ch.isalpha():
            shift = ord(key[k % len(key)].lower()) - ord("a")
            out.append(caesar(ch, -shift if decrypt else shift))
            k += 1
        else:
            out.append(ch)
    return "".join(out)

msg = "Hello, World! xyz"
enc = caesar(msg, 3)
print(enc, caesar(enc, -3) == msg, caesar(msg, 26) == msg, caesar("abc", 13))
v = vigenere("Attack at Dawn", "lemon")
print(v, vigenere(v, "lemon", True))
def is_palindrome(s):
    cleaned = [c.lower() for c in s if c.isalnum()]
    return cleaned == cleaned[::-1]
def anagrams(a, b):
    return sorted(a.replace(" ", "").lower()) == sorted(b.replace(" ", "").lower())
print(is_palindrome("A man, a plan, a canal: Panama"), is_palindrome("nope"), anagrams("Dormitory", "dirty room"), anagrams("a", "b"))
print("".join(sorted(set("mississippi"))), "mississippi".count("ss"), "mississippi".replace("ss", "SS"), "mississippi".title())
