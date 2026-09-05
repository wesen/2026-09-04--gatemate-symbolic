package syntax

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestLexSpansAndTrivia(t *testing.T) {
	source := "// λ🙂\r\ndef main : Int = n-1 <= 2;"
	tokens, ds := Lex(context.Background(), source)
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	want := []TokenKind{Def, Ident, Colon, TInt, Assign, Ident, Minus, Integer, LessEqual, Integer, Semicolon, EOF}
	kinds := []TokenKind{}
	for _, tok := range tokens {
		kinds = append(kinds, tok.Kind)
		if tok.Text != source[tok.Span.Start:tok.Span.End] {
			t.Fatalf("bad token slice %+v", tok)
		}
	}
	if !reflect.DeepEqual(kinds, want) {
		t.Fatal(kinds)
	}
	if tokens[0].Span.Start != len("// λ🙂\r\n") {
		t.Fatal(tokens[0])
	}
}
func TestLexKeywordsAndOperators(t *testing.T) {
	src := "def fun let rec in if then else case of Nil Cons true false Int Bool ListInt ( ) { } : ; , = -> + - * == <= _x name2"
	tokens, ds := Lex(context.Background(), src)
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	if len(tokens) != 34 {
		t.Fatalf("got %d tokens", len(tokens))
	}
	for _, tok := range tokens[:len(tokens)-3] {
		if tok.Kind == Ident || tok.Kind == Invalid {
			t.Fatal(tok)
		}
	}
}
func TestLexInvalidAndBounds(t *testing.T) {
	for _, src := range []string{"@", "λ", string([]byte{255}), "//" + string([]byte{255})} {
		_, ds := Lex(context.Background(), src)
		if len(ds) != 1 {
			t.Fatalf("%q: %+v", src, ds)
		}
	}
	_, ds := Lex(context.Background(), strings.Repeat("@", 100))
	if len(ds) != MaxDiagnostics {
		t.Fatal(len(ds))
	}
	_, ds = Lex(context.Background(), strings.Repeat(" ", MaxSourceBytes+1))
	if ds[0].Code != "SOURCE_LIMIT" {
		t.Fatal(ds)
	}
	_, ds = Lex(context.Background(), strings.Repeat("x ", MaxTokens+1))
	if ds[0].Code != "TOKEN_LIMIT" {
		t.Fatal(ds)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, ds = Lex(ctx, "x")
	if ds[0].Code != "CANCELED" {
		t.Fatal(ds)
	}
}
