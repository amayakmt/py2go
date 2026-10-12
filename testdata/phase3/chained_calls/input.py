def words() -> list[str]:
    return "alpha,beta,gamma".split(",")


print("a,b".split(",")[0].upper())
print(words()[1].upper().lower()[::-1])
print(len(words()), len(sorted(words(), reverse=True)[0]))
words().append("x")
print(" ".join(sorted(words())).strip().split()[-1])
print(str(len(words())) + "/" + str(max(len("ab"), len("abc"))))
print(words()[0][0], words()[-1][-1])
nested: list[list[str]] = [words(), ["z"]]
print(nested[0][2][1:3], nested[1][0] * 2)
