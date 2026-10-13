def caesar(s, k):
    out = []
    for ch in s:
        if "a" <= ch <= "z":
            out.append(chr((ord(ch) - 97 + k) % 26 + 97))
        elif "A" <= ch <= "Z":
            out.append(chr((ord(ch) - 65 + k) % 26 + 65))
        else:
            out.append(ch)
    return "".join(out)

def vigenere(s, key, decrypt=False):
    out, i = [], 0
    for ch in s:
        if ch.isalpha() and ch.isascii():
            base = 65 if ch.isupper() else 97
            shift = ord(key[i % len(key)].lower()) - 97
            if decrypt:
                shift = -shift
            out.append(chr((ord(ch) - base + shift) % 26 + base))
            i += 1
        else:
            out.append(ch)
    return "".join(out)

msg = "The Quick Brown Fox Jumps Over The Lazy Dog! Ünïcödé stays."
for k in (0, 3, 13, 25, -5, 100):
    e = caesar(msg, k)
    print(k, e, caesar(e, -k) == msg)
v = vigenere(msg, "lemon")
print(v, vigenere(v, "lemon", True) == msg)
freq = {}
for ch in msg.lower():
    if ch.isalpha():
        freq[ch] = freq.get(ch, 0) + 1
print(sorted(freq.items(), key=lambda kv: (-kv[1], kv[0]))[:6], len(freq))
print(msg.translate(str.maketrans("aeiou", "AEIOU")), "".join(sorted(set(msg.lower()) - set(" !."))), msg.encode("ascii", "replace").decode(), msg.encode("ascii", "ignore").decode()[-10:])
