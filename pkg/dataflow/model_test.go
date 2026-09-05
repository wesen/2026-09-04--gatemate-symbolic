package dataflow

import (
	"context"
	"math"
	"math/rand"
	"testing"
)

func TestValuesAndEnvelope(t *testing.T) {
	cases := []struct {
		op         Opcode
		a, b, want Value
	}{
		{Mul, Int(7), Int(6), Int(42)}, {Mul, Int(-32768), Int(-32768), Int(1073741824)},
		{Mul, Int(32768), Int(1), ErrorValue(IntegerOverflow)}, {Add, Int(math.MaxInt32), Int(1), ErrorValue(IntegerOverflow)},
		{Sub, Int(math.MinInt32), Int(1), ErrorValue(IntegerOverflow)}, {Sub, Int(1), Int(-2), Int(3)},
		{Less, Int(-2), Int(1), Bool(true)}, {BoolToInt, Bool(true), 0, Int(1)},
		{BoolToInt, Int(1), 0, ErrorValue(BadTag)}, {BoolToInt, Value(1)<<36 | 2, 0, ErrorValue(BadTag)},
		{Add, Bool(false), Int(2), ErrorValue(BadTag)}, {Copy, Bool(true), 0, Bool(true)},
		{Copy, Value(1) << 32, 0, ErrorValue(BadTag)}, {Opcode(99), Int(0), Int(0), ErrorValue(BadDescriptor)},
	}
	for _, tc := range cases {
		if got := Evaluate(tc.op, tc.a, tc.b); got != tc.want {
			t.Fatalf("op %d: got %010X want %010X", tc.op, got, tc.want)
		}
	}
	r := rand.New(rand.NewSource(17))
	for i := 0; i < 1000; i++ {
		x, y := int32(r.Intn(65536)-32768), int32(r.Intn(65536)-32768)
		if got := Evaluate(Mul, Int(x), Int(y)); got != Int(x*y) {
			t.Fatal(x, y, got)
		}
		token := Token{byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(64)), byte(r.Intn(2)), r.Intn(2) == 1, byte(r.Intn(256)), byte(r.Intn(256)), Int(x)}
		p, err := token.Bytes()
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeToken(p[:])
		if err != nil || got != token {
			t.Fatal(got, token, err)
		}
	}
	if _, err := (Token{Node: 64}).Bytes(); err == nil {
		t.Fatal("wide node accepted")
	}
	if _, err := DecodeToken(make([]byte, 9)); err == nil {
		t.Fatal("short token accepted")
	}
}

func directed() []Token {
	a := Inputs(0, 0, [6]int32{7, 6, 3, 5, 2, 9})
	b := Inputs(1, 0, [6]int32{10, -2, 4, 8, 7, 1})
	return []Token{a[0], b[5], a[3], b[0], a[5], b[1], a[1], b[4], a[2], b[2], a[4], b[3]}
}
func run(t *testing.T, m *Transaction, input []Token, seed int64) []Token {
	t.Helper()
	r := rand.New(rand.NewSource(seed))
	index := 0
	var outputs []Token
	for cycle := 0; cycle < 20000; cycle++ {
		if index < len(input) && r.Intn(3) != 0 && m.Inject(input[index]) {
			index++
		}
		if err := m.Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
		if r.Intn(4) == 0 {
			if out := m.Poll(); out != nil {
				outputs = append(outputs, *out)
			}
		}
		if index == len(input) && m.Quiescent() {
			return outputs
		}
	}
	t.Fatalf("did not drain: %+v", m.Metrics)
	return nil
}
func newModel(t *testing.T, c Config) *Transaction {
	t.Helper()
	m, err := NewTransaction(c)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func assertOutputs(t *testing.T, got []Token, want map[byte]int32) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("outputs %+v want %+v", got, want)
	}
	seen := map[byte]bool{}
	for _, o := range got {
		v, ok := want[o.Context]
		if !ok || seen[o.Context] || !o.Final || o.Value.Tag() != 0 || o.Value.Int32() != v {
			t.Fatalf("bad output %+v want %+v", o, want)
		}
		seen[o.Context] = true
	}
}
func TestDirectedAcrossLatencyAndCapacity(t *testing.T) {
	var semantic Semantic
	var expected []Token
	for _, input := range directed() {
		expected = append(expected, semantic.Deliver(input)...)
	}
	assertOutputs(t, expected, map[byte]int32{0: 58, 1: 12})
	for _, latency := range []int{1, 2, 4, 8} {
		for _, depth := range []int{1, 2, 8} {
			for seed := int64(1); seed <= 20; seed++ {
				m := newModel(t, Config{depth, depth, depth, latency, 8})
				out := run(t, m, directed(), seed)
				assertOutputs(t, out, map[byte]int32{0: 58, 1: 12})
				if m.Metrics.Activations != 12 || m.Metrics.Mul != 4 || m.Metrics.ALU != 8 {
					t.Fatal(m.Metrics)
				}
				seen := map[[3]byte]bool{}
				for _, a := range m.Trace {
					key := [3]byte{a.Context, a.Epoch, a.Node}
					if seen[key] {
						t.Fatal("duplicate activation", key)
					}
					seen[key] = true
				}
				if m.Metrics.InputHigh > depth || m.Metrics.CompletionHigh > depth || m.Metrics.OutputHigh > depth || m.Metrics.ReadyHigh > Contexts*Nodes {
					t.Fatal("capacity", m.Metrics)
				}
			}
		}
	}
}
func TestRandomValuesAgainstIndependentExpression(t *testing.T) {
	r := rand.New(rand.NewSource(1234))
	for iteration := 0; iteration < 100; iteration++ {
		var input []Token
		want := map[byte]int32{}
		for c := byte(0); c < Contexts; c++ {
			var v [6]int32
			for i := range v {
				v[i] = int32(r.Intn(2001) - 1000)
			}
			want[c] = v[0]*v[1] + v[2]*v[3]
			if v[4] < v[5] {
				want[c]++
			}
			input = append(input, Inputs(c, 0, v)...)
		}
		r.Shuffle(len(input), func(i, j int) { input[i], input[j] = input[j], input[i] })
		m := newModel(t, Config{2, 1, 1, 8, 8})
		assertOutputs(t, run(t, m, input, int64(iteration)), want)
	}
}
func TestCancellationDropsInflightMultiplyAndReusesEpoch(t *testing.T) {
	m := newModel(t, DefaultConfig())
	m.Config.MulLatency = 8
	m.mul = make([]*Token, 8)
	old := Inputs(2, 0, [6]int32{7, 6, 3, 5, 2, 9})
	m.Inject(old[0])
	m.Inject(old[1])
	for m.Metrics.Mul == 0 {
		_ = m.Tick(context.Background())
	}
	if !m.Cancel(2) || m.EpochFor(2) != 1 {
		t.Fatal("cancel refused")
	}
	input := append([]Token{old[2]}, Inputs(2, 1, [6]int32{10, -2, 4, 8, 7, 1})...)
	out := run(t, m, input, 42)
	assertOutputs(t, out, map[byte]int32{2: 12})
	if out[0].Epoch != 1 || m.Metrics.StaleCompletion != 1 || m.Metrics.StaleInput != 1 {
		t.Fatal(out, m.Metrics)
	}
}
func TestOutputStabilityCancellationAndWrap(t *testing.T) {
	m := newModel(t, Config{8, 8, 1, 4, 2})
	input := Inputs(0, 0, [6]int32{7, 6, 3, 5, 2, 9})
	for _, in := range input {
		m.Inject(in)
	}
	for i := 0; i < 200; i++ {
		_ = m.Tick(context.Background())
	}
	if len(m.output) != 1 {
		t.Fatal("no held output")
	}
	held := m.output[0]
	if m.Cancel(0) {
		t.Fatal("withdrew offered output")
	}
	for i := 0; i < 100; i++ {
		_ = m.Tick(context.Background())
		if m.output[0] != held {
			t.Fatal("unstable output")
		}
	}
	if m.Poll() == nil || !m.Cancel(0) || !m.Cancel(0) || !m.Cancel(0) || m.EpochFor(0) != 3 {
		t.Fatal("epochs")
	}
	m.Inject(Source(1, 0, 0, 0, Int(1)))
	if m.Cancel(0) {
		t.Fatal("wrapped while input in flight")
	}
	_ = m.Tick(context.Background())
	if !m.Cancel(0) || m.EpochFor(0) != 0 {
		t.Fatal("quiescent wrap refused")
	}
}
func TestCopyFanoutAndIsolatedErrors(t *testing.T) {
	copyInput := []Token{Source(0, 0, 6, 0, Int(7)), Source(0, 0, 0, 1, Int(6)), Source(0, 0, 1, 1, Int(5)), Source(0, 0, 3, 0, Int(2)), Source(0, 0, 3, 1, Int(9))}
	m := newModel(t, Config{1, 1, 1, 8, 8})
	assertOutputs(t, run(t, m, copyInput, 55), map[byte]int32{0: 78})
	if m.Metrics.Activations != 7 {
		t.Fatal(m.Metrics)
	}
	for _, faultCase := range []struct {
		input []Token
		code  byte
	}{
		{[]Token{Source(0, 0, 0, 0, Int(1)), Source(0, 0, 0, 0, Int(2))}, DuplicateOperand},
		{[]Token{Source(0, 0, 4, 1, Int(1))}, BadDestination},
		{[]Token{Source(0, 0, 0, 0, Bool(true)), Source(0, 0, 0, 1, Int(2))}, BadTag},
		{[]Token{Source(0, 0, 0, 0, Int(32768)), Source(0, 0, 0, 1, Int(1))}, IntegerOverflow},
	} {
		m := newModel(t, Config{2, 1, 1, 8, 8})
		out := run(t, m, append(faultCase.input, Inputs(1, 0, [6]int32{10, -2, 4, 8, 7, 1})...), 7)
		if len(out) != 2 {
			t.Fatal(out)
		}
		for _, v := range out {
			if v.Context == 0 {
				if v.Value != ErrorValue(faultCase.code) || v.Subtype != faultCase.code {
					t.Fatal(v)
				}
			} else if v.Context != 1 || v.Value != Int(12) {
				t.Fatal(v)
			}
		}
	}
}
func TestTickCancellationAndInvalidConfig(t *testing.T) {
	if _, err := NewTransaction(Config{}); err == nil {
		t.Fatal("zero capacities")
	}
	m := newModel(t, DefaultConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := m.Tick(ctx); err == nil || m.Metrics.Cycles != 0 {
		t.Fatal("cancelled tick mutated")
	}
	m.Inject(Token{Context: 255})
	_ = m.Tick(context.Background())
	if m.Metrics.Invalid != 1 {
		t.Fatal("invalid context indexed")
	}
}
