package dataflowide

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/pkg/errors"
	df "github.com/wesen/2026-09-04--gatemate-symbolic/pkg/dataflow"
)

const MaxSourceBytes = 128 * 1024

type Action struct {
	Kind    string     `json:"kind"`
	Context byte       `json:"context,omitempty"`
	Values  []int32    `json:"values,omitempty"`
	Token   *df.Token  `json:"token,omitempty"`
	Ticks   uint32     `json:"ticks,omitempty"`
	Value   *df.Value  `json:"value,omitempty"`
	Epoch   *byte      `json:"epoch,omitempty"`
	Config  *df.Config `json:"config,omitempty"`
}
type Scenario struct {
	Version     int      `json:"version"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Actions     []Action `json:"actions"`
}
type Diagnostic struct {
	Action  int    `json:"action"`
	Message string `json:"message"`
}

func ParseScenario(source string) (Scenario, []Diagnostic) {
	var s Scenario
	if len(source) > MaxSourceBytes {
		return s, []Diagnostic{{-1, "source exceeds 128 KiB"}}
	}
	dec := json.NewDecoder(strings.NewReader(source))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		return s, []Diagnostic{{-1, err.Error()}}
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return s, []Diagnostic{{-1, "expected one JSON object"}}
	}
	d := []Diagnostic{}
	add := func(n int, message string) { d = append(d, Diagnostic{n, message}) }
	if s.Version != 1 {
		add(-1, "version must be 1")
	}
	if len(strings.TrimSpace(s.Name)) == 0 || len(s.Name) > 80 {
		add(-1, "name must contain 1..80 characters")
	}
	if len(s.Description) > 2000 {
		add(-1, "description exceeds 2000 characters")
	}
	if len(s.Actions) == 0 || len(s.Actions) > 512 {
		add(-1, "actions must contain 1..512 entries")
		return s, d
	}
	if s.Actions[0].Kind != "reset" {
		add(0, "a scenario must start with an explicit reset")
	}
	var totalTicks uint64
	for n, a := range s.Actions {
		switch a.Kind {
		case "inputs":
			if a.Context >= df.Contexts || len(a.Values) != 6 {
				add(n, "inputs requires context 0..3 and six signed 32-bit values")
			}
		case "expect":
			if a.Context >= df.Contexts || a.Value == nil || uint64(*a.Value)>>40 != 0 {
				add(n, "expect requires context 0..3 and a 40-bit encoded value")
			}
		default:
			if err := (df.Operation{Kind: a.Kind, Context: a.Context, Token: a.Token, Ticks: a.Ticks, Config: a.Config}).Validate(); err != nil {
				add(n, err.Error())
			}
		}
		totalTicks += uint64(a.Ticks)
		if a.Config != nil && (a.Kind != "reset" || a.Config.InputDepth > 8 || a.Config.CompletionDepth > 8 || a.Config.OutputDepth > 8) {
			add(n, "configuration is limited to reset with queue depths 1..8")
		}
	}
	if totalTicks > 2_000_000 {
		add(-1, "scenario exceeds two million explicit ticks")
	}
	return s, d
}
func FormatScenario(s Scenario) string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b) + "\n"
}
func validateScenario(s Scenario) error {
	_, d := ParseScenario(FormatScenario(s))
	if len(d) > 0 {
		return errors.Errorf("action %d: %s", d[0].Action, d[0].Message)
	}
	return nil
}
func value(v df.Value) *df.Value { return &v }
func Examples() map[string]string {
	examples := map[string]Scenario{
		"book": {1, "Interleaved expressions", "Two contexts evaluate a*b + c*d + int(e<f). Results are 58 and 12.", []Action{
			{Kind: "reset"}, {Kind: "inputs", Context: 0, Values: []int32{7, 6, 3, 5, 2, 9}}, {Kind: "inputs", Context: 1, Values: []int32{10, -2, 4, 8, 7, 1}},
			{Kind: "tick", Ticks: 96}, {Kind: "poll"}, {Kind: "poll"}, {Kind: "expect", Context: 0, Value: value(58)}, {Kind: "expect", Context: 1, Value: value(12)}}},
		"copy": {1, "COPY fanout", "Node 6 supplies the same value to both multipliers.", []Action{
			{Kind: "reset"}, {Kind: "inject", Token: token(df.Source(0, 0, 6, 0, 7))}, {Kind: "inject", Token: token(df.Source(0, 0, 0, 1, 6))}, {Kind: "inject", Token: token(df.Source(0, 0, 1, 1, 5))},
			{Kind: "inject", Token: token(df.Source(0, 0, 3, 0, 2))}, {Kind: "inject", Token: token(df.Source(0, 0, 3, 1, 9))}, {Kind: "tick", Ticks: 96}, {Kind: "poll"}, {Kind: "expect", Value: value(78)}}},
		"fault": {1, "Isolated operand fault", "A duplicate input closes context 0; context 1 still completes.", []Action{
			{Kind: "reset"}, {Kind: "inject", Token: token(df.Source(0, 0, 0, 0, 7))}, {Kind: "inject", Token: token(df.Source(0, 0, 0, 0, 8))},
			{Kind: "inputs", Context: 1, Values: []int32{10, -2, 4, 8, 7, 1}}, {Kind: "tick", Ticks: 96}, {Kind: "poll"}, {Kind: "poll"},
			{Kind: "expect", Value: value(df.ErrorValue(df.DuplicateOperand))}, {Kind: "expect", Context: 1, Value: value(12)}}},
		"cancel": {1, "Cancel an occupied multiplier", "Stop after six ticks, inspect the multiply stages, then cancel context 2 and submit epoch-one work.", []Action{
			{Kind: "reset"}, {Kind: "inject", Token: token(df.Source(2, 0, 0, 0, 7))}, {Kind: "inject", Token: token(df.Source(2, 0, 0, 1, 6))}, {Kind: "tick", Ticks: 6}, {Kind: "cancel", Context: 2},
			{Kind: "inject", Token: token(df.Source(2, 0, 1, 0, 3))}, {Kind: "inputs", Context: 2, Values: []int32{10, -2, 4, 8, 7, 1}}, {Kind: "tick", Ticks: 96}, {Kind: "poll"}, {Kind: "expect", Context: 2, Value: value(12)}}},
	}
	out := map[string]string{}
	for id, s := range examples {
		out[id] = FormatScenario(s)
	}
	return out
}
func token(t df.Token) *df.Token          { return &t }
func canonicalScenario(s Scenario) []byte { b, _ := json.Marshal(s); return bytes.TrimSpace(b) }
