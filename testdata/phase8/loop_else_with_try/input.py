log: list[str] = []
for i in range(4):
    try:
        if i == 2:
            break
        log.append("body " + str(i))
    finally:
        log.append("fin " + str(i))
else:
    log.append("else")
print(log)
log.clear()
for i in range(2):
    try:
        log.append("body " + str(i))
    finally:
        log.append("fin " + str(i))
else:
    log.append("else")
print(log)
