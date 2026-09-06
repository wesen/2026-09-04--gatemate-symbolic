// Package serial implements the checked LFL1 UART transport. Each public method
// serializes a complete operation; uncertain failures invalidate synchronization.
package serial

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/compile"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/ir"
	"github.com/wesen/2026-09-04--gatemate-symbolic/pkg/lazylang/machine"
	uart "go.bug.st/serial"
)

type Port interface {
	io.ReadWriteCloser
	ResetInputBuffer() error
}
type Client struct {
	mu                   sync.Mutex
	port                 Port
	synchronized, closed bool
	artifactID           string
	// Capture may be set before the first operation to retain UART evidence.
	Capture io.Writer
}

func Open(path string) (*Client, error) {
	p, err := uart.Open(path, &uart.Mode{BaudRate: 115200, DataBits: 8, Parity: uart.NoParity, StopBits: uart.OneStopBit})
	if err != nil {
		return nil, errors.Wrap(err, "open LFL1 UART")
	}
	if err = p.SetReadTimeout(50 * time.Millisecond); err != nil {
		_ = p.Close()
		return nil, errors.Wrap(err, "set UART timeout")
	}
	return &Client{port: p}, nil
}
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return c.port.Close()
}
func checksum(b []byte) byte {
	var sum byte
	for _, v := range b {
		sum ^= v
	}
	return sum
}
func encode(command byte, payload []byte) string {
	if len(payload) == 0 {
		return string(command) + "\n"
	}
	return fmt.Sprintf("%c%s%02x\n", command, hex.EncodeToString(payload), checksum(payload))
}
func (c *Client) exchange(ctx context.Context, command byte, payload []byte) (line string, err error) {
	if c.closed {
		return "", errors.New("UART closed")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	defer func() {
		if err != nil {
			c.synchronized = false
		}
	}()
	if err = ctx.Err(); err != nil {
		return "", err
	}
	request := encode(command, payload)
	if c.Capture != nil {
		_, _ = fmt.Fprintf(c.Capture, "> %s", request)
	}
	pending := []byte(request)
	for len(pending) > 0 {
		if err = ctx.Err(); err != nil {
			return "", err
		}
		var n int
		n, err = c.port.Write(pending)
		if err != nil {
			return "", errors.Wrap(err, "write LFL1 request")
		}
		if n <= 0 {
			return "", io.ErrShortWrite
		}
		pending = pending[n:]
	}
	response := []byte{}
	buffer := make([]byte, 1)
	for len(response) < 40 {
		if err = ctx.Err(); err != nil {
			return "", errors.Wrap(err, "read LFL1 response")
		}
		var n int
		n, err = c.port.Read(buffer)
		if err != nil {
			return "", errors.Wrap(err, "read UART")
		}
		if n == 0 {
			continue
		}
		response = append(response, buffer[0])
		if buffer[0] == '\n' {
			line = string(response)
			if c.Capture != nil {
				_, _ = fmt.Fprintf(c.Capture, "< %s", line)
			}
			return line, nil
		}
	}
	return "", errors.New("oversized LFL1 response")
}
func (c *Client) ack(ctx context.Context, command byte, payload []byte) error {
	line, err := c.exchange(ctx, command, payload)
	if err != nil {
		return err
	}
	if line != "A\n" {
		c.synchronized = false
		return errors.Errorf("LFL1 %c rejected: %q", command, line)
	}
	return nil
}
func record(line string, kind byte) ([16]byte, error) {
	var out [16]byte
	if len(line) != 36 || line[0] != kind || line[35] != '\n' {
		return out, errors.Errorf("invalid %c record: %q", kind, line)
	}
	raw, err := hex.DecodeString(line[1:35])
	if err != nil {
		return out, errors.Wrap(err, "decode LFL1 record")
	}
	if checksum(raw) != 0 {
		return out, errors.New("LFL1 response checksum mismatch")
	}
	copy(out[:], raw[:16])
	return out, nil
}
func u16(n uint16) []byte { return []byte{byte(n >> 8), byte(n)} }
func (c *Client) query(ctx context.Context, page uint16) ([16]byte, error) {
	line, err := c.exchange(ctx, 'Q', u16(page))
	if err != nil {
		return [16]byte{}, err
	}
	out, err := record(line, 'S')
	if err != nil {
		c.synchronized = false
	}
	return out, err
}
func (c *Client) reset(ctx context.Context) error {
	c.synchronized = false
	c.artifactID = ""
	if err := c.port.ResetInputBuffer(); err != nil {
		return errors.Wrap(err, "reset UART input")
	}
	if err := c.ack(ctx, 'R', nil); err != nil {
		return err
	}
	sig, err := c.query(ctx, 0)
	if err != nil {
		return err
	}
	if hex.EncodeToString(sig[:]) != "4c464c31010800020008000040000000" {
		return errors.Errorf("unexpected LFL1 profile: %x", sig)
	}
	c.synchronized = true
	return nil
}
func (c *Client) Reset(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reset(ctx)
}
func (c *Client) ready() error {
	if !c.synchronized {
		return errors.New("UART state uncertain: reset required")
	}
	return nil
}

// Load validates before resetting the device, writes contiguous memories,
// verifies every word, and commits only after readback matches.
func (c *Client) Load(ctx context.Context, a *compile.Artifact) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if a == nil {
		return errors.New("missing artifact")
	}
	if err = a.Validate(); err != nil {
		return errors.Wrap(err, "validate artifact before device mutation")
	}
	defer func() {
		if err != nil {
			c.synchronized = false
		}
	}()
	if err = c.reset(ctx); err != nil {
		return err
	}
	payload := append(u16(uint16(len(a.Code))), u16(uint16(len(a.Heap)))...)
	payload = append(payload, u16(a.Root)...)
	payload = append(payload, u16(a.ConstantEnd)...)
	if err = c.ack(ctx, 'B', payload); err != nil {
		return err
	}
	for _, part := range []struct {
		command byte
		words   []string
		page    uint16
	}{{'C', a.Code, 0x2000}, {'H', a.Heap, 0x1000}} {
		for i, word := range part.words {
			raw, _ := hex.DecodeString(word)
			if err = c.ack(ctx, part.command, append(u16(uint16(i)), raw...)); err != nil {
				return err
			}
		}
	}
	for i, span := range a.Provenance {
		if err = c.ack(ctx, 'V', append(u16(uint16(i)), u16(span)...)); err != nil {
			return err
		}
	}
	for _, part := range []struct {
		words []string
		page  uint16
	}{{a.Code, 0x2000}, {a.Heap, 0x1000}} {
		for i, word := range part.words {
			raw, _ := hex.DecodeString(word)
			var got [16]byte
			got, err = c.query(ctx, part.page+uint16(i))
			if err != nil {
				return err
			}
			padded := make([]byte, 16)
			copy(padded[16-len(raw):], raw)
			if !bytes.Equal(got[:], padded) {
				return errors.Errorf("LFL1 readback mismatch at %04x", part.page+uint16(i))
			}
		}
	}
	for i, span := range a.Provenance {
		var got [16]byte
		got, err = c.query(ctx, 0x4000+uint16(i))
		if err != nil {
			return err
		}
		var want [16]byte
		binary.BigEndian.PutUint16(want[14:], span)
		if got != want {
			return errors.Errorf("provenance mismatch at %d", i)
		}
	}
	if err = c.ack(ctx, 'K', nil); err != nil {
		return err
	}
	c.artifactID = a.ID
	return nil
}
func (c *Client) Demand(ctx context.Context, ref uint16) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ready(); err != nil {
		return err
	}
	return c.ack(ctx, 'F', u16(ref))
}
func (c *Client) Tick(ctx context.Context, n uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ready(); err != nil {
		return err
	}
	if n > 1000000 {
		return errors.New("tick budget exceeds 1000000")
	}
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, n)
	return c.ack(ctx, 'T', payload)
}
func (c *Client) Poll(ctx context.Context) (uint16, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ready(); err != nil {
		return ir.Absent, false, err
	}
	line, err := c.exchange(ctx, 'P', nil)
	if err != nil {
		return ir.Absent, false, err
	}
	if line == "N\n" {
		return ir.Absent, false, nil
	}
	data, err := record(line, 'O')
	if err != nil {
		c.synchronized = false
		return ir.Absent, false, err
	}
	if !bytes.Equal(data[:14], make([]byte, 14)) {
		c.synchronized = false
		return ir.Absent, false, errors.New("noncanonical output reference")
	}
	return binary.BigEndian.Uint16(data[14:]), true, nil
}
func (c *Client) Snapshot(ctx context.Context) (out machine.Snapshot, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.ready(); err != nil {
		return out, err
	}
	defer func() {
		if err != nil {
			c.synchronized = false
		}
	}()
	pages := make([][16]byte, 5)
	for i := 1; i < 5; i++ {
		pages[i], err = c.query(ctx, uint16(i))
		if err != nil {
			return out, err
		}
	}
	p := pages[1]
	get := func(b []byte) uint16 { return binary.BigEndian.Uint16(b) }
	out = machine.Snapshot{ArtifactID: c.artifactID, State: machine.State(p[0]), Valid: p[1]&1 != 0, CurrentRef: get(p[2:4]), CodeID: get(p[4:6]), EnvRef: get(p[6:8]), CommittedTop: get(p[10:12]), ResultRef: get(p[12:14]), Root: get(pages[2][0:2]), ConstantEnd: get(pages[2][6:8]), IndSteps: get(pages[2][8:10]), EnvSteps: get(pages[2][10:12]), Heap: []string{}, Provenance: []uint16{}, Stack: []string{}, Trace: []machine.Event{}}
	stackCount := get(p[8:10])
	traceCount := get(pages[3][0:2])
	p = pages[4]
	out.Allocation = machine.Allocation{Active: pages[1][1]&8 != 0, Base: get(p[0:2]), ReservedEnd: get(p[2:4]), ObjectCount: get(p[4:6]), NextWrite: get(p[6:8]), Kind: p[8], Stage: p[9]}
	end := out.CommittedTop
	if out.Allocation.Active {
		end = out.Allocation.ReservedEnd
	}
	if end > ir.Capacity || stackCount > ir.StackCapacity || traceCount > 64 {
		return out, errors.New("invalid device snapshot dimensions")
	}
	for i := uint16(0); i < end; i++ {
		var h, v [16]byte
		h, err = c.query(ctx, 0x1000+i)
		if err != nil {
			return out, err
		}
		v, err = c.query(ctx, 0x4000+i)
		if err != nil {
			return out, err
		}
		out.Heap = append(out.Heap, hex.EncodeToString(h[6:]))
		out.Provenance = append(out.Provenance, get(v[14:]))
	}
	for i := uint16(0); i < stackCount; i++ {
		var f [16]byte
		f, err = c.query(ctx, 0x3000+i)
		if err != nil {
			return out, err
		}
		out.Stack = append(out.Stack, hex.EncodeToString(f[:]))
	}
	for i := uint16(0); i < traceCount; i++ {
		var hi, lo [16]byte
		hi, err = c.query(ctx, 0x5000+2*i)
		if err != nil {
			return out, err
		}
		lo, err = c.query(ctx, 0x5001+2*i)
		if err != nil {
			return out, err
		}
		raw := append(hi[:], lo[:]...)
		out.Trace = append(out.Trace, machine.Event{Cycle: binary.BigEndian.Uint32(raw[:4]), Address: get(raw[4:6]), Span: get(raw[6:8]), Kind: raw[8], Old: hex.EncodeToString(raw[12:22]), New: hex.EncodeToString(raw[22:32])})
	}
	cnt := &out.Counters
	targets := []*uint32{&cnt.Cycles, &cnt.CodeReads, &cnt.HeapReads, &cnt.Allocations, &cnt.AllocatedThunks, &cnt.AllocatedFunctions, &cnt.AllocatedEnvironments, &cnt.AllocatedCons, &cnt.AllocatedIntegers, &cnt.Claims, &cnt.Updates, &cnt.Adds, &cnt.Subtracts, &cnt.Multiplies, &cnt.Equalities, &cnt.ComparisonsLE, &cnt.Indirections, &cnt.EnvironmentSteps, &cnt.MaxStack, &cnt.MaxHeap, &cnt.OutputStalls, &cnt.Faults, &cnt.TraceDropped}
	for i, target := range targets {
		var data [16]byte
		data, err = c.query(ctx, uint16(16+i))
		if err != nil {
			return out, err
		}
		*target = binary.BigEndian.Uint32(data[12:])
	}
	return out, nil
}

// IsRejected reports a known protocol rejection for diagnostics. It does not
// restore synchronization; the caller must reset after a failed operation.
func IsRejected(err error) bool { return err != nil && strings.Contains(err.Error(), "rejected:") }

// ReadObject inspects one committed heap slot without demanding it.
func (c *Client) ReadObject(ctx context.Context, ref uint16) (ir.Object, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ready(); err != nil {
		return ir.Object{}, err
	}
	if ref >= ir.Capacity {
		return ir.Object{}, errors.New("heap reference exceeds profile")
	}
	data, err := c.query(ctx, 0x1000+ref)
	if err != nil {
		return ir.Object{}, err
	}
	return ir.ParseObject(hex.EncodeToString(data[6:]))
}
