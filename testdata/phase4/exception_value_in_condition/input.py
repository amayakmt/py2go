attempts: int = 0
done: bool = False
while not done:
    attempts += 1
    try:
        if attempts < 3:
            raise ValueError("try again")
        done = True
        print("succeeded after", attempts)
    except ValueError as e:
        print("attempt", attempts, "failed:", e)
