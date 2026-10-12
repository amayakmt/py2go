def describe(name: str, age: int, city: str = "nowhere") -> str:
    return name + " " + str(age) + " " + city


print(describe("ann", 30))
print(describe("ann", 30, "paris"))
print(describe(name="bob", age=25))
print(describe(age=41, name="cy", city="rome"))
print(describe("di", city="oslo", age=19))
print(describe("ed", 50, city="lima"))
