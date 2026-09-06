package ir_test

import (
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
	"testing"
)

func TestPackedRecords(t *testing.T) {
	o := ir.Object{Tag: ir.Thunk, A: 0x1234, B: 0xabcd, Payload: 0x89abcdef}
	if o.Hex() != "05001234abcd89abcdef" {
		t.Fatal(o.Hex())
	}
	if got, err := ir.ParseObject(o.Hex()); err != nil || got != o {
		t.Fatal(got, err)
	}
	c := ir.Code{Op: ir.Case, A: 0x1234, B: 0xabcd, C: 0x5678, Immediate: 0x12345678, Span: 0x4321}
	if c.Hex() != "09001234abcd56781234567843210000" {
		t.Fatal(c.Hex())
	}
	if got, err := ir.ParseCode(c.Hex()); err != nil || got != c {
		t.Fatal(got, err)
	}
	if _, err := ir.ParseObject("00"); err == nil {
		t.Fatal("short object accepted")
	}
	if _, err := ir.ParseCode("zz"); err == nil {
		t.Fatal("bad code accepted")
	}
	f := ir.Frame{Kind: ir.Update, A: 0x1234, Span: 0x5678}
	if f.Hex() != "06001234000000000000000056780000" {
		t.Fatal(f.Hex())
	}
}
