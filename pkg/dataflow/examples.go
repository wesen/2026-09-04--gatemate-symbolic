package dataflow

import (
	"context"
	"github.com/pkg/errors"
)

// RunExample checks end-to-end behavior on either the physical device or model.
// Every run starts with an explicitly requested reset and bounded execution.
func RunExample(ctx context.Context, e Engine, name string) ([]Token, Snapshot, error) {
	fail := func(err error) ([]Token, Snapshot, error) { return nil, Snapshot{}, err }
	if _, err := e.Execute(ctx, Operation{Kind: "reset"}); err != nil {
		return fail(err)
	}
	inject := func(t Token) error {
		for attempts := 0; attempts < 1000; attempts++ {
			_, err := e.Execute(ctx, Operation{Kind: "inject", Token: &t})
			if err == nil {
				return nil
			}
			if !errors.Is(err, ErrFull) {
				return err
			}
			if _, err = e.Execute(ctx, Operation{Kind: "tick", Ticks: 1}); err != nil {
				return err
			}
		}
		return errors.New("input did not gain credit")
	}
	want := map[byte]Value{}
	var inputs []Token
	switch name {
	case "book":
		a, b := Inputs(0, 0, [6]int32{7, 6, 3, 5, 2, 9}), Inputs(1, 0, [6]int32{10, -2, 4, 8, 7, 1})
		inputs = []Token{a[0], b[5], a[3], b[0], a[5], b[1], a[1], b[4], a[2], b[2], a[4], b[3]}
		want[0] = Int(58)
		want[1] = Int(12)
	case "copy":
		inputs = []Token{Source(0, 0, 6, 0, 7), Source(0, 0, 0, 1, 6), Source(0, 0, 1, 1, 5), Source(0, 0, 3, 0, 2), Source(0, 0, 3, 1, 9)}
		want[0] = Int(78)
	case "fault":
		inputs = []Token{Source(0, 0, 0, 0, 7), Source(0, 0, 0, 0, 8)}
		inputs = append(inputs, Inputs(1, 0, [6]int32{10, -2, 4, 8, 7, 1})...)
		want[0] = ErrorValue(DuplicateOperand)
		want[1] = Int(12)
	case "cancel":
		if err := inject(Source(2, 0, 0, 0, 7)); err != nil {
			return fail(err)
		}
		if err := inject(Source(2, 0, 0, 1, 6)); err != nil {
			return fail(err)
		}
		found := false
		for i := 0; i < 32; i++ {
			if _, err := e.Execute(ctx, Operation{Kind: "tick", Ticks: 1}); err != nil {
				return fail(err)
			}
			s, err := e.Snapshot(ctx)
			if err != nil {
				return fail(err)
			}
			for _, t := range s.Mul {
				if t != nil {
					found = true
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fail(errors.New("multiply never became visible in flight"))
		}
		if _, err := e.Execute(ctx, Operation{Kind: "cancel", Context: 2}); err != nil {
			return fail(err)
		}
		inputs = append([]Token{Source(2, 0, 1, 0, 3)}, Inputs(2, 1, [6]int32{10, -2, 4, 8, 7, 1})...)
		want[2] = Int(12)
	default:
		return fail(errors.New("unknown example"))
	}
	for _, t := range inputs {
		if err := inject(t); err != nil {
			return fail(err)
		}
	}
	outputs := []Token{}
	seen := map[byte]bool{}
	for i := 0; i < 100; i++ {
		if _, err := e.Execute(ctx, Operation{Kind: "tick", Ticks: 16}); err != nil {
			return fail(err)
		}
		for j := 0; j < 16; j++ {
			t, err := e.Execute(ctx, Operation{Kind: "poll"})
			if err != nil {
				return fail(err)
			}
			if t == nil {
				break
			}
			value, ok := want[t.Context]
			if !ok || seen[t.Context] || t.Value != value || !t.Final {
				return fail(errors.Errorf("unexpected result %+v", *t))
			}
			seen[t.Context] = true
			outputs = append(outputs, *t)
		}
		if len(outputs) == len(want) {
			s, err := e.Snapshot(ctx)
			if err != nil {
				return fail(err)
			}
			if s.Quiescent {
				return outputs, s, nil
			}
		}
	}
	return fail(errors.New("example did not drain within 1600 enabled cycles"))
}
