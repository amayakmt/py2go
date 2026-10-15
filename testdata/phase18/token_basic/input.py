import token

print(token.NAME, token.NUMBER, token.OP, token.ENDMARKER, token.NEWLINE, token.INDENT)
print(token.tok_name[token.NAME], token.tok_name[token.OP], len(token.tok_name))
print(token.ISTERMINAL(token.NAME), token.ISNONTERMINAL(token.NAME), token.ISEOF(token.ENDMARKER))
print(token.EXACT_TOKEN_TYPES["+"], token.EXACT_TOKEN_TYPES["->"], token.tok_name[token.EXACT_TOKEN_TYPES["**="]])
print(token.LPAR, token.RPAR, token.COLONEQUAL, token.FSTRING_START, token.N_TOKENS)
print(sorted(token.__all__)[:6])
