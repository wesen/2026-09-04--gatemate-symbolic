from pathlib import Path
p=Path('pkg/lazylang/syntax/lexer.go');s=p.read_text();needle='func Lex(ctx context.Context, source string) ([]Token, []Diagnostic) {\n'
s=s.replace(needle,needle+'\tif err := ctx.Err(); err != nil {\n\t\treturn []Token{{Kind: EOF, Span: Span{0, 0}}}, []Diagnostic{{"CANCELED", err.Error(), Span{0, 0}}}\n\t}\n')
p.write_text(s)
