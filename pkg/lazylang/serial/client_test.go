package serial

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
)

type fakePort struct {
	response bytes.Buffer
	requests []string
	memory   map[uint16][16]byte
	corrupt  bool
}

var _ Port = (*fakePort)(nil)

func (p *fakePort) Close() error               { return nil }
func (p *fakePort) ResetInputBuffer() error    { p.response.Reset(); return nil }
func (p *fakePort) Read(b []byte) (int, error) { return p.response.Read(b) }
func responseRecord(kind byte, data [16]byte) string {
	return fmt.Sprintf("%c%x%02x\n", kind, data, checksum(data[:]))
}
func (p *fakePort) Write(b []byte) (int, error) {
	line := string(b)
	p.requests = append(p.requests, line)
	payload := []byte{}
	if len(line) > 2 {
		raw, err := hex.DecodeString(line[1 : len(line)-1])
		if err != nil || checksum(raw) != 0 {
			panic("bad request")
		}
		payload = raw[:len(raw)-1]
	}
	reply := "A\n"
	switch line[0] {
	case 'R':
		p.memory = map[uint16][16]byte{}
		sig, _ := hex.DecodeString("4c464c31010800020008000040000000")
		var page [16]byte
		copy(page[:], sig)
		p.memory[0] = page
	case 'C', 'H', 'V':
		page := map[byte]uint16{'C': 0x2000, 'H': 0x1000, 'V': 0x4000}[line[0]] + binary.BigEndian.Uint16(payload[:2])
		var data [16]byte
		copy(data[18-len(payload):], payload[2:])
		p.memory[page] = data
	case 'Q':
		page := binary.BigEndian.Uint16(payload)
		data := p.memory[page]
		if p.corrupt && page >= 0x2000 {
			data[0] ^= 1
		}
		reply = responseRecord('S', data)
	case 'P':
		reply = "N\n"
	}
	p.response.WriteString(reply)
	return len(b), nil
}
func TestLoadReadbackAndRejectBeforeMutation(t *testing.T) {
	a, ds := compile.Compile(context.Background(), `def main : Int = 42;`)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	p := &fakePort{}
	c := &Client{port: p}
	if err := c.Load(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	if !c.synchronized || c.artifactID != a.ID || p.requests[len(p.requests)-1] != "K\n" {
		t.Fatal("load did not commit")
	}
	before := len(p.requests)
	a.Source += " "
	if err := c.Load(context.Background(), a); err == nil {
		t.Fatal("corrupt artifact accepted")
	}
	if len(p.requests) != before {
		t.Fatal("invalid artifact mutated device")
	}
	if _, ok, err := c.Poll(context.Background()); err != nil || ok {
		t.Fatal(ok, err)
	}
}
func TestReadbackFailureRequiresReset(t *testing.T) {
	a, ds := compile.Compile(context.Background(), `def main : Int = 42;`)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	p := &fakePort{corrupt: true}
	c := &Client{port: p}
	if err := c.Load(context.Background(), a); err == nil || !strings.Contains(err.Error(), "readback mismatch") {
		t.Fatal(err)
	}
	for _, request := range p.requests {
		if request == "K\n" {
			t.Fatal("committed corrupt readback")
		}
	}
	before := len(p.requests)
	if err := c.Demand(context.Background(), a.Root); err == nil {
		t.Fatal("accepted unsynchronized demand")
	}
	if len(p.requests) != before {
		t.Fatal("sent unsynchronized demand")
	}
	p.corrupt = false
	if err := c.Load(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Tick(ctx, 1); err == nil {
		t.Fatal("ignored cancellation")
	}
	if c.synchronized {
		t.Fatal("failed command still synchronized")
	}
}
func TestRecords(t *testing.T) {
	var data [16]byte
	data[15] = 25
	if got, err := record(responseRecord('O', data), 'O'); err != nil || got != data {
		t.Fatal(got, err)
	}
	for _, line := range []string{"O\n", responseRecord('S', data), "O0000000000000000000000000000001900\n", "Ozz00000000000000000000000000001919\n"} {
		if _, err := record(line, 'O'); err == nil {
			t.Fatal("bad record accepted", line)
		}
	}
	if encode('F', []byte{0, 15}) != "F000f0f\n" {
		t.Fatal("request encoding")
	}
}

func TestSnapshotDecoding(t *testing.T) {
	a, ds := compile.Compile(context.Background(), `def main : Int = 42;`)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	p := &fakePort{}
	c := &Client{port: p}
	if err := c.Load(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	var status, meta, trace [16]byte
	status[0] = 18
	status[1] = 5
	binary.BigEndian.PutUint16(status[10:12], uint16(len(a.Heap)))
	binary.BigEndian.PutUint16(status[12:14], 13)
	binary.BigEndian.PutUint16(meta[0:2], a.Root)
	binary.BigEndian.PutUint16(meta[6:8], a.ConstantEnd)
	binary.BigEndian.PutUint16(trace[0:2], 1)
	p.memory[1] = status
	p.memory[2] = meta
	p.memory[3] = trace
	for i := uint16(0); i < 23; i++ {
		var value [16]byte
		binary.BigEndian.PutUint32(value[12:], uint32(i))
		p.memory[16+i] = value
	}
	raw := make([]byte, 32)
	binary.BigEndian.PutUint32(raw[:4], 9)
	binary.BigEndian.PutUint16(raw[4:6], 14)
	binary.BigEndian.PutUint16(raw[6:8], 3)
	raw[8] = 2
	raw[12] = 5
	raw[22] = 7
	var hi, lo [16]byte
	copy(hi[:], raw[:16])
	copy(lo[:], raw[16:])
	p.memory[0x5000] = hi
	p.memory[0x5001] = lo
	snap, err := c.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snap.Root != a.Root || snap.ConstantEnd != a.ConstantEnd || snap.ResultRef != 13 || !snap.Valid || len(snap.Heap) != len(a.Heap) || snap.Counters.TraceDropped != 22 {
		t.Fatalf("%+v", snap)
	}
	if len(snap.Trace) != 1 || snap.Trace[0].Cycle != 9 || snap.Trace[0].Old != "05000000000000000000" || snap.Trace[0].New != "07000000000000000000" {
		t.Fatal(snap.Trace)
	}
}
