# Lazy language syntax

This package implements the lexer and handwritten parser for GATEMATE-SYMBOLIC-010. It returns a source-preserving AST for later binding resolution, monomorphic type checking and compilation. Those later passes are not implemented here.

```go
program, diagnostics := syntax.Parse(ctx, source)
if len(diagnostics) != 0 {
    // Display each diagnostic's code, message and half-open byte span.
    // Complete later definitions may still be present for editor tooling.
    return
}
for _, definition := range program.Definitions {
    // Inspect definition.Name, Annotation and Value.
}
```

`Parse(context.Context, string) (*Program, []Diagnostic)` parses one or more `def name : Type = expression;` definitions. It omits malformed definitions and returns diagnostics in source order. A recovered program with diagnostics must not be executed. Duplicate definitions, free variables, `main` validation and type mismatches belong to the next semantic passes. For example, parsing an annotated expression does not establish that its annotation is correct.

`Lex(context.Context, string) ([]Token, []Diagnostic)` exposes tokens including EOF. Token text is an exact slice of the source. Identifiers use ASCII letters, digits and underscore, starting with a letter or underscore. Keywords are case-sensitive. Whitespace includes spaces, tabs, LF and CR; `//` comments may contain Unicode. Invalid UTF-8 and unsupported characters produce diagnostics. Unsupported non-comment characters also produce Invalid tokens to prevent accidentally joining expressions around the error.

All spans use inclusive Start and exclusive End UTF-8 byte offsets into the original source, including for EOF. The browser will need a byte-to-UTF-16 mapping before applying these ranges to an editor. Identifier and annotation spans are retained separately from enclosing nodes. GroupExpr preserves parentheses and their full range.

## Grammar and precedence

The parser accepts Int, Bool, ListInt, grouped types and right-associative function types. Lambda parameters and local/top-level bindings require annotations. The expression forms are integer/boolean literals, names, lambdas, application, let/let rec, arithmetic, comparisons, if, Nil, Cons and two-branch list case.

| Form | Binding |
|---|---|
| Application by adjacency | 40, left-associative |
| Multiplication | 30, left-associative |
| Addition and subtraction | 20, left-associative |
| Equality and less-than-or-equal | 10, cannot chain |

`f x + y * 2` parses as `(f x) + (y * 2)`. `f x y` parses as `(f x) y`. Compound arguments such as lambdas, lets, ifs and cases need parentheses. Lambdas and branch/binding bodies otherwise extend right to the relevant delimiter.

The lexer emits minus separately. In prefix position the parser accepts minus followed by an integer token, including -2147483648. In infix position it is subtraction, so `n-1` and `n - 1` agree. Write `f (-1)` to pass a negative argument; `f -1` means subtraction. Unary negation of an arbitrary expression is not supported. Integer literals outside signed int32 produce INTEGER_RANGE diagnostics.

Comparisons such as `a <= b == c` are rejected. Explicit grouping, such as `(a <= b) == c`, is syntactically accepted; the future checker decides whether its operand types are valid.

Constructors are fully saturated: `Cons(head, tail)`. Case syntax is deliberately fixed:

```text
case xs of {
  Nil -> emptyBranch;
  Cons(head, tail) -> nonemptyBranch
}
```

There is no trailing semicolon after the Cons branch. The enclosing top-level definition still ends with its own semicolon. Pattern names remain unresolved syntax, including duplicate names to be rejected by the future binding pass.

## Diagnostics and bounds

Recovery scans from the failed definition's beginning, counts braces, and stops at a top-level semicolon or a new `def`. This preserves later definitions after malformed nested cases and missing closing braces. The new-definition anchor also handles a missing semicolon. Parsing must always advance or reach EOF.

The current bounds are one MiB of source, 65,536 non-EOF tokens, 64 diagnostics, and 256 recursive parsing frames. Expression-tree height is independently bounded at 256, covering flat application/operator chains as well as visibly nested syntax. Lexer source/token truncation returns diagnostics without attempting to interpret the truncated stream as a complete program. Cancellation is reported as CANCELED and includes a source position.

Diagnostic codes include INVALID_CHARACTER, INVALID_UTF8, SOURCE_LIMIT, TOKEN_LIMIT, CANCELED, EXPECTED_DEFINITION, EXPECTED_TOKEN, EXPECTED_TYPE, EXPECTED_EXPRESSION, INTEGER_RANGE, CHAINED_COMPARISON and NESTING_LIMIT. Messages describe the local error; clients should use the stable code and span rather than parse message wording.

## Validation and examples

```sh
go test ./pkg/lazylang/syntax -count=1
go test -race ./pkg/lazylang/syntax -count=1
go test ./pkg/lazylang/syntax -run '^$' -fuzz '^FuzzParse$' -fuzztime 15s -parallel 2
```

The six files under examples/lazylang cover shared arithmetic, closure capture, an unused divergent argument, cyclic demand, a productive recursive list and eight squares. Current tests establish that they parse with complete ASTs and valid spans. Their intended evaluation results will be tested when the independent semantic evaluator exists.

The fuzz test checks deterministic parsing, bounded diagnostics, token/source slicing and nested AST/type spans for arbitrary input, including malformed UTF-8. The structural tests inspect operator association, application shape, recursive constructs, signed integer boundaries, recovery and cancellation.
