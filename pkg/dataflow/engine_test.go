package dataflow

import (
	"context"
	"encoding/binary"
	"io"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestRecordFrames(t *testing.T) {
	token := Source(2, 7, 3, 1, Int(-32))
	p, _ := token.Bytes()
	line := EncodeRequest('O', p[:])
	got, err := DecodeRecord(line, 'O')
	if err != nil || got != p {
		t.Fatal(got, err)
	}
	for _, bad := range []string{line[:23], "S" + line[1:], line[:21] + "00\n", line[:4] + "zz" + line[6:]} {
		if _, err := DecodeRecord(bad, 'O'); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}
func TestSnapshotDecodeAndCompleteness(t *testing.T) {
	pages := map[byte][10]byte{}
	for _, r := range [][2]byte{{0, 13}, {16, 25}, {32, 59}, {64, 91}, {96, 103}, {112, 119}, {128, 135}, {144, 147}} {
		for a := r[0]; a <= r[1]; a++ {
			pages[a] = [10]byte{}
		}
	}
	pages[0] = [10]byte{1, 4, 7, 4, 8, 8, 8, 8, 0, 0}
	pages[1] = [10]byte{3, 2, 1, 0, 0x22, 1, 1, 0, 0, 0}
	tok := Source(2, 2, 3, 1, Int(-2))
	wire, _ := tok.Bytes()
	pages[96] = wire
	pair := [10]byte{}
	binary.BigEndian.PutUint32(pair[1:5], 42)
	binary.BigEndian.PutUint32(pair[6:10], 99)
	pages[32] = pair
	pages[64] = [10]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 13}
	pages[145] = wire
	s, err := DecodeSnapshot(pages)
	if err != nil {
		t.Fatal(err)
	}
	if s.Epochs != [4]byte{0, 1, 2, 3} || !s.Closed[1] || s.Closed[0] || s.Input[0] != tok || s.Errors[1] == nil || s.Slots[0].Values != [2]Value{42, 0} || !s.Slots[0].Pending {
		t.Fatalf("bad snapshot %+v", s)
	}
	delete(pages, 44)
	if _, err := DecodeSnapshot(pages); err == nil {
		t.Fatal("accepted incomplete pages")
	}
}
func TestSnapshotDetachedAndOperations(t *testing.T) {
	ctx := context.Background()
	m := newModel(t, DefaultConfig())
	tok := Source(0, 0, 0, 0, Int(7))
	if _, err := m.Execute(ctx, Operation{Kind: "inject", Token: &tok}); err != nil {
		t.Fatal(err)
	}
	s, _ := m.Snapshot(ctx)
	s.Input[0].Value = 900
	s.Counters["source"] = 99
	again, _ := m.Snapshot(ctx)
	if again.Input[0].Value != 7 || again.Counters["source"] != 1 {
		t.Fatal("snapshot aliases engine")
	}
	for _, o := range []Operation{{Kind: "inject"}, {Kind: "cancel", Context: 4}, {Kind: "tick", Ticks: 1_000_001}, {Kind: "wat"}} {
		if _, err := m.Execute(ctx, o); err == nil {
			t.Fatal("accepted invalid operation")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := m.Execute(canceled, Operation{Kind: "reset"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

type scriptedPort struct {
	responses []string
	input     string
	writes    []string
	closed    bool
}

func (p *scriptedPort) Write(b []byte) (int, error) {
	p.writes = append(p.writes, string(b))
	if len(p.responses) == 0 {
		return 0, io.ErrUnexpectedEOF
	}
	p.input = p.responses[0]
	p.responses = p.responses[1:]
	return len(b), nil
}
func (p *scriptedPort) Read(b []byte) (int, error) {
	if len(p.input) == 0 {
		return 0, io.EOF
	}
	b[0] = p.input[0]
	p.input = p.input[1:]
	return 1, nil
}
func (p *scriptedPort) Close() error            { p.closed = true; return nil }
func (p *scriptedPort) ResetInputBuffer() error { p.input = ""; return nil }

var _ wirePort = &scriptedPort{}

func TestSerialRejectionAndUncertainDelivery(t *testing.T) {
	p := &scriptedPort{responses: []string{"!04\n", "!05\n", "garbage\n"}}
	s := &Serial{port: p, synchronized: true}
	ctx := context.Background()
	tok := Source(0, 0, 0, 0, 7)
	if _, err := s.Execute(ctx, Operation{Kind: "inject", Token: &tok}); !errors.Is(err, ErrFull) || !s.synchronized {
		t.Fatal(err)
	}
	if _, err := s.Execute(ctx, Operation{Kind: "cancel"}); !errors.Is(err, ErrBlocked) || !s.synchronized {
		t.Fatal(err)
	}
	if _, err := s.Execute(ctx, Operation{Kind: "tick", Ticks: 1}); err == nil || s.synchronized {
		t.Fatal("malformed ACK retained synchronization")
	}
	before := len(p.writes)
	if _, err := s.Execute(ctx, Operation{Kind: "tick", Ticks: 1}); err == nil || !strings.Contains(err.Error(), "reset required") {
		t.Fatal(err)
	}
	if len(p.writes) != before {
		t.Fatal("retried uncertain mutation")
	}
	if err := s.Close(); err != nil || !p.closed {
		t.Fatal(err)
	}
}
