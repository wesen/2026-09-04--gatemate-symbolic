// Package syntax parses the lazy functional laboratory language. It preserves
// source names and UTF-8 byte spans; binding resolution and typing are later passes.
package syntax

// Span is a half-open byte range in the original source, including trivia between
// a node's first and last tokens. It is not a UTF-16 editor position.
type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}
type Diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Span    Span   `json:"span"`
}
type Identifier struct {
	Name string `json:"name"`
	Span Span   `json:"span"`
}
type Program struct {
	Definitions []Definition `json:"definitions"`
	Span        Span         `json:"span"`
}
type Definition struct {
	Name       Identifier
	Annotation *Type
	Value      Expr
	Span       Span
}
type TypeKind string

const (
	IntType      TypeKind = "Int"
	BoolType     TypeKind = "Bool"
	ListIntType  TypeKind = "ListInt"
	FunctionType TypeKind = "function"
)

type Type struct {
	Kind              TypeKind
	Parameter, Result *Type
	Span              Span
}
type Parameter struct {
	Name       Identifier
	Annotation *Type
	Span       Span
}

// Expr is the common expression-node interface. GroupExpr preserves explicit
// parentheses, including the distinction between chained and grouped comparisons.
type Expr interface {
	SourceSpan() Span
	expression()
}
type Node struct{ Range Span }

func (n Node) SourceSpan() Span { return n.Range }
func (Node) expression()        {}

type IntExpr struct {
	Node
	Value int32
}
type BoolExpr struct {
	Node
	Value bool
}
type VarExpr struct {
	Node
	Name Identifier
}
type LambdaExpr struct {
	Node
	Parameter Parameter
	Body      Expr
}
type ApplyExpr struct {
	Node
	Function, Argument Expr
}
type LetExpr struct {
	Node
	Recursive   bool
	Name        Identifier
	Annotation  *Type
	Value, Body Expr
}
type BinaryExpr struct {
	Node
	Operator    TokenKind
	Left, Right Expr
}
type IfExpr struct {
	Node
	Condition, Then, Else Expr
}
type NilExpr struct{ Node }
type ConsExpr struct {
	Node
	Head, Tail Expr
}
type CaseExpr struct {
	Node
	Scrutinee, NilBranch Expr
	Head, Tail           Identifier
	ConsBranch           Expr
}
type GroupExpr struct {
	Node
	Inner Expr
}
