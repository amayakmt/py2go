scores: dict[str, int] = {"bob": 85, "ann": 92, "cy": 78, "dee": 92}
print(sorted(scores))
print(sorted(scores, reverse=True))
print(sorted(scores.items()))
print(sorted(scores.values()), sum(scores.values()), max(scores.values()), min(scores.values()))
print(max(scores), min(scores), len(scores))
print(sum(scores.values()) / len(scores))
