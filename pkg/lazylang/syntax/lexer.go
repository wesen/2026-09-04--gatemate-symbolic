package syntax

import (
	"context"
	"fmt"
	"unicode/utf8"
)

const MaxSourceBytes = 1 << 20
const MaxTokens = 65536
const MaxDiagnostics = 64
const MaxNesting = 256

type TokenKind string

const (
	EOF       TokenKind = "end of input"
	Invalid   TokenKind = "invalid token"
	Ident     TokenKind = "identifier"
	Integer   TokenKind = "integer"
	Def       TokenKind = "def"
	Fun       TokenKind = "fun"
	Let       TokenKind = "let"
	Rec       TokenKind = "rec"
	In        TokenKind = "in"
	If        TokenKind = "if"
	Then      TokenKind = "then"
	Else      TokenKind = "else"
	Case      TokenKind = "case"
	Of        TokenKind = "of"
	Nil       TokenKind = "Nil"
	Cons      TokenKind = "Cons"
	True      TokenKind = "true"
	False     TokenKind = "false"
	TInt      TokenKind = "Int"
	TBool     TokenKind = "Bool"
	TListInt  TokenKind = "ListInt"
	LParen    TokenKind = "("
	RParen    TokenKind = ")"
	LBrace    TokenKind = "{"
	RBrace    TokenKind = "}"
	Colon     TokenKind = ":"
	Semicolon TokenKind = ";"
	Comma     TokenKind = ","
	Assign    TokenKind = "="
	Arrow     TokenKind = "->"
	Plus      TokenKind = "+"
	Minus     TokenKind = "-"
	Star      TokenKind = "*"
	Equal     TokenKind = "=="
	LessEqual TokenKind = "<="
)

type Token struct {
	Kind TokenKind
	Text string
	Span Span
}

var keywords = map[string]TokenKind{"def": Def, "fun": Fun, "let": Let, "rec": Rec, "in": In, "if": If, "then": Then, "else": Else, "case": Case, "of": Of, "Nil": Nil, "Cons": Cons, "true": True, "false": False, "Int": TInt, "Bool": TBool, "ListInt": TListInt}

func letter(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b == '_' }
func digit(b byte) bool  { return b >= '0' && b <= '9' }

// Lex returns tokens including EOF and bounded diagnostics. Invalid characters
// produce Invalid tokens so the parser cannot silently join expressions across them.
func Lex(ctx context.Context, source string) ([]Token, []Diagnostic) {
	tokens := make([]Token, 0)
	diagnostics := make([]Diagnostic, 0)
	add := func(code, message string, span Span) {
		if len(diagnostics) < MaxDiagnostics {
			diagnostics = append(diagnostics, Diagnostic{code, message, span})
		}
	}
	if len(source) > MaxSourceBytes {
		return []Token{{Kind: EOF, Span: Span{0, 0}}}, []Diagnostic{{"SOURCE_LIMIT", "source exceeds one MiB", Span{0, len(source)}}}
	}
	i := 0
	for i < len(source) {
		if err := ctx.Err(); err != nil {
			add("CANCELED", err.Error(), Span{i, i})
			break
		}
		if len(tokens) == MaxTokens {
			add("TOKEN_LIMIT", "source exceeds token limit", Span{i, i})
			break
		}
		start := i
		b := source[i]
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			i++
			continue
		}
		if b == '/' && i+1 < len(source) && source[i+1] == '/' {
			i += 2
			for i < len(source) && source[i] != '\n' {
				_, size := utf8.DecodeRuneInString(source[i:])
				if size == 1 && source[i] >= 128 {
					add("INVALID_UTF8", "invalid UTF-8 in comment", Span{i, i + 1})
				}
				i += size
			}
			continue
		}
		kind := Invalid
		switch {
		case letter(b):
			i++
			for i < len(source) && (letter(source[i]) || digit(source[i])) {
				i++
			}
			kind = Ident
			if k, ok := keywords[source[start:i]]; ok {
				kind = k
			}
		case digit(b):
			i++
			for i < len(source) && digit(source[i]) {
				i++
			}
			kind = Integer
		default:
			i++
			switch b {
			case '(':
				kind = LParen
			case ')':
				kind = RParen
			case '{':
				kind = LBrace
			case '}':
				kind = RBrace
			case ':':
				kind = Colon
			case ';':
				kind = Semicolon
			case ',':
				kind = Comma
			case '+':
				kind = Plus
			case '*':
				kind = Star
			case '-':
				kind = Minus
				if i < len(source) && source[i] == '>' {
					i++
					kind = Arrow
				}
			case '=':
				kind = Assign
				if i < len(source) && source[i] == '=' {
					i++
					kind = Equal
				}
			case '<':
				if i < len(source) && source[i] == '=' {
					i++
					kind = LessEqual
				}
			}
			if kind == Invalid {
				_, size := utf8.DecodeRuneInString(source[start:])
				i = start + size
				if size == 1 && b >= 128 {
					add("INVALID_UTF8", "invalid UTF-8", Span{start, i})
				} else {
					add("INVALID_CHARACTER", fmt.Sprintf("unexpected character %q", source[start:i]), Span{start, i})
				}
			}
		}
		tokens = append(tokens, Token{kind, source[start:i], Span{start, i}})
	}
	tokens = append(tokens, Token{Kind: EOF, Span: Span{i, i}})
	return tokens, diagnostics
}
