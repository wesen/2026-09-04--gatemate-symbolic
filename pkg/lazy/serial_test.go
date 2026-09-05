package lazy

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestWireEncoding(t *testing.T) {
	if got := request('W', []byte{0, 3, 0x30, 0, 2, 0, 0}); got != "W0003300002000031\n" {
		t.Fatal(got)
	}
	if request('F', []byte{0, 6}) != "F000606\n" {
		t.Fatal("force framing")
	}
	p := make([]byte, 10)
	p[9] = 42
	line := fmt.Sprintf("S%X2A\n", p)
	got, e := decodeRecord(line, 'S')
	if e != nil || got[9] != 42 {
		t.Fatal(e)
	}
	if _, e = decodeRecord(strings.Replace(line, "2A\n", "2B\n", 1), 'S'); e == nil {
		t.Fatal("checksum accepted")
	}
}

type fakePort struct {
	response []byte
	writes   []string
	closed   bool
}

func (p *fakePort) Read(b []byte) (int, error) {
	if len(p.response) == 0 {
		return 0, io.EOF
	}
	n := copy(b, p.response)
	p.response = p.response[n:]
	return n, nil
}
func (p *fakePort) Write(b []byte) (int, error) {
	p.writes = append(p.writes, string(b))
	return len(b), nil
}
func (p *fakePort) Close() error            { p.closed = true; return nil }
func (p *fakePort) ResetInputBuffer() error { return nil }
func TestSerialUncertainResponseRequiresReset(t *testing.T) {
	p := &fakePort{response: []byte("bad\n")}
	s := &Serial{port: p, synchronized: true}
	if _, e := s.Execute(context.Background(), Operation{Kind: "tick", Ticks: 1}); e == nil {
		t.Fatal("bad ack accepted")
	}
	if _, e := s.Execute(context.Background(), Operation{Kind: "force"}); e == nil {
		t.Fatal("uncertain state accepted")
	}
	if len(p.writes) != 1 {
		t.Fatal("sent after uncertainty")
	}
	_ = s.Close()
	if !p.closed {
		t.Fatal("port not closed")
	}
}
