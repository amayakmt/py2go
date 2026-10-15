total = 0
for i in range(3000):
    for j in range(3000):
        total += (i * j) % 7
print(total)
