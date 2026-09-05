package lazy

import (
	"context"
	"github.com/pkg/errors"
)

func RunExample(ctx context.Context, e Engine, name string) ([]Word, Snapshot, error) {
	i, err := Example(name)
	if err != nil {
		return nil, Snapshot{}, err
	}
	if _, err = e.Execute(ctx, Operation{Kind: "load", Image: &i}); err != nil {
		return nil, Snapshot{}, err
	}
	if _, err = e.Execute(ctx, Operation{Kind: "force", Root: i.Root}); err != nil {
		return nil, Snapshot{}, err
	}
	if _, err = e.Execute(ctx, Operation{Kind: "tick", Ticks: 20000}); err != nil {
		return nil, Snapshot{}, err
	}
	s, err := e.Snapshot(ctx)
	if err != nil {
		return nil, s, err
	}
	want, heap := Reference(i, StackCapacity)
	if !s.Valid || s.Result != want {
		return nil, s, errors.New("example output differs from reference")
	}
	for n, w := range heap {
		if s.Heap[n] != w {
			return nil, s, errors.New("memoized heap differs from reference")
		}
	}
	if name == "shared" && (s.Counters.Claims != 1 || s.Counters.Updates != 1 || s.Counters.Muls != 1) {
		return nil, s, errors.New("shared thunk did not execute exactly once")
	}
	v, err := e.Execute(ctx, Operation{Kind: "poll"})
	if err != nil {
		return nil, s, err
	}
	if v == nil || *v != want {
		return nil, s, errors.New("poll differs from offered output")
	}
	return []Word{*v}, s, nil
}
