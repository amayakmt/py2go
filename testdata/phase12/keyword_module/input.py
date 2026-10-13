import keyword
print(keyword.iskeyword("for"), keyword.iskeyword("print"), keyword.iskeyword("None"), keyword.iskeyword("match"), keyword.issoftkeyword("match"), keyword.issoftkeyword("for"))
print(len(keyword.kwlist), keyword.kwlist[:5], keyword.kwlist[-3:], sorted(keyword.kwlist) == keyword.kwlist, keyword.softkwlist)
print([w for w in "def class foo while bar lambda".split() if keyword.iskeyword(w)])
