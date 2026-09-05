package lazy

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.bug.st/serial"
)

type wirePort interface {
	io.ReadWriteCloser
	ResetInputBuffer() error
}
type Serial struct {
	mu                   sync.Mutex
	port                 wirePort
	synchronized, closed bool
	// Set Capture before the first operation. It records request/response lines.
	Capture io.Writer
}

var _ Engine = &Serial{}

func NewSerial(path string) (*Serial, error) {
	p, err := serial.Open(path, &serial.Mode{BaudRate: 115200, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		return nil, errors.Wrap(err, "open lazy UART")
	}
	if err := p.SetReadTimeout(50 * time.Millisecond); err != nil {
		_ = p.Close()
		return nil, errors.Wrap(err, "set UART read timeout")
	}
	return &Serial{port: p}, nil
}
func (s *Serial) exchange(ctx context.Context, command string) (line string, err error) {
	if s.closed {
		return "", errors.New("serial device closed")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	defer func() {
		if err != nil {
			s.synchronized = false
		}
	}()
	if err = ctx.Err(); err != nil {
		return "", err
	}
	if s.Capture != nil {
		_, _ = fmt.Fprintf(s.Capture, "> %s", command)
	}
	data := []byte(command)
	for len(data) > 0 {
		if err = ctx.Err(); err != nil {
			return "", err
		}
		var n int
		n, err = s.port.Write(data)
		if err != nil {
			return "", errors.Wrap(err, "write lazy request")
		}
		if n <= 0 {
			return "", io.ErrShortWrite
		}
		data = data[n:]
	}
	buffer := make([]byte, 32)
	response := []byte{}
	for {
		if err = ctx.Err(); err != nil {
			return "", errors.Wrap(err, "read lazy response")
		}
		n, e := s.port.Read(buffer)
		if e != nil {
			return "", errors.Wrap(e, "read lazy UART")
		}
		response = append(response, buffer[:n]...)
		if len(response) > 32 {
			return "", errors.New("response exceeds framing bound")
		}
		if i := strings.IndexByte(string(response), '\n'); i >= 0 {
			if i != len(response)-1 {
				return "", errors.New("extra bytes after UART response")
			}
			line = string(response)
			if s.Capture != nil {
				_, _ = fmt.Fprintf(s.Capture, "< %s", line)
			}
			if strings.HasPrefix(line, "!") {
				return "", errors.Errorf("device rejected request: %s", strings.TrimSpace(line))
			}
			return line, nil
		}
	}
}

func request(kind byte, p []byte) string {
	if len(p) == 0 {
		return string(kind) + "\n"
	}
	sum := byte(0)
	for _, v := range p {
		sum ^= v
	}
	return fmt.Sprintf("%c%X%02X\n", kind, p, sum)
}
func decodeRecord(line string, kind byte) ([]byte, error) {
	if len(line) != 24 || line[0] != kind || line[23] != '\n' {
		return nil, errors.New("invalid record framing")
	}
	b, e := hex.DecodeString(line[1:23])
	if e != nil {
		return nil, e
	}
	sum := byte(0)
	for _, v := range b {
		sum ^= v
	}
	if sum != 0 {
		return nil, errors.New("invalid response checksum")
	}
	return b[:10], nil
}
func (s *Serial) ack(ctx context.Context, kind byte, p []byte) error {
	line, e := s.exchange(ctx, request(kind, p))
	if e != nil {
		return e
	}
	if line != "A\n" {
		s.synchronized = false
		return errors.New("expected acknowledgment")
	}
	return nil
}
func (s *Serial) query(ctx context.Context, page uint16) ([]byte, error) {
	line, e := s.exchange(ctx, request('Q', []byte{byte(page >> 8), byte(page)}))
	if e != nil {
		return nil, e
	}
	b, e := decodeRecord(line, 'S')
	if e != nil {
		s.synchronized = false
	}
	return b, e
}
func readWord(b []byte) Word { return Word(b[0])<<32 | Word(binary.BigEndian.Uint32(b[1:5])) }
func (s *Serial) reset(ctx context.Context) error {
	if !s.synchronized {
		timer := time.NewTimer(250 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
		if e := s.port.ResetInputBuffer(); e != nil {
			return e
		}
	}
	if e := s.ack(ctx, 'R', nil); e != nil {
		return e
	}
	b, e := s.query(ctx, 0)
	if e != nil {
		return e
	}
	if string(b[:4]) != "LAZY" || b[4] != 1 || binary.BigEndian.Uint16(b[5:7]) != HeapCapacity || binary.BigEndian.Uint16(b[7:9]) != StackCapacity {
		s.synchronized = false
		return errors.New("unsupported lazy reducer capabilities")
	}
	s.synchronized = true
	return nil
}
func (s *Serial) Execute(ctx context.Context, o Operation) (*Word, error) {
	if e := o.Validate(); e != nil {
		return nil, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("engine closed")
	}
	if o.Kind == "reset" {
		return nil, s.reset(ctx)
	}
	if o.Kind == "load" {
		if e := s.reset(ctx); e != nil {
			return nil, e
		}
		for n, w := range o.Image.Nodes {
			p := []byte{byte(n >> 8), byte(n), byte(w >> 32), byte(w >> 24), byte(w >> 16), byte(w >> 8), byte(w)}
			if e := s.ack(ctx, 'W', p); e != nil {
				return nil, e
			}
		}
		for n, w := range o.Image.Nodes {
			p, e := s.query(ctx, uint16(0x1000+n))
			if e != nil {
				return nil, e
			}
			if readWord(p[5:]) != w {
				s.synchronized = false
				return nil, errors.New("heap readback mismatch")
			}
		}
		return nil, nil
	}
	if !s.synchronized {
		return nil, errors.New("UART state uncertain; explicit reset required")
	}
	switch o.Kind {
	case "force":
		return nil, s.ack(ctx, 'F', []byte{byte(o.Root >> 8), byte(o.Root)})
	case "tick":
		p := make([]byte, 4)
		binary.BigEndian.PutUint32(p, o.Ticks)
		return nil, s.ack(ctx, 'T', p)
	case "poll":
		line, e := s.exchange(ctx, "P\n")
		if e != nil {
			return nil, e
		}
		if line == "N\n" {
			return nil, nil
		}
		p, e := decodeRecord(line, 'O')
		if e != nil {
			s.synchronized = false
			return nil, e
		}
		w := readWord(p[5:])
		return &w, nil
	}
	return nil, errors.New("unknown operation")
}
func (s *Serial) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.port.Close()
}
func (s *Serial) Snapshot(ctx context.Context) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := Snapshot{Source: "serial", Heap: []Word{}, Stack: []Frame{}, Trace: []Mutation{}}
	if !s.synchronized || s.closed {
		return out, errors.New("UART requires reset or is closed")
	}
	p, e := s.query(ctx, 1)
	if e != nil {
		return out, e
	}
	states := []string{"idle", "fetch", "heap wait", "evaluate", "return", "stack wait", "return dispatch", "update wait", "update", "multiply", "output"}
	if int(p[0]) >= len(states) {
		s.synchronized = false
		return out, errors.New("invalid controller state")
	}
	out.State = states[p[0]]
	out.Valid = p[1] != 0
	out.Address = binary.BigEndian.Uint16(p[2:4])
	heapSize, stackSize := binary.BigEndian.Uint16(p[4:6]), binary.BigEndian.Uint16(p[6:8])
	if heapSize > HeapCapacity || stackSize > StackCapacity {
		return out, errors.New("invalid device capacities")
	}
	p, e = s.query(ctx, 2)
	if e != nil {
		return out, e
	}
	out.Result = readWord(p[5:])
	p, e = s.query(ctx, 3)
	if e != nil {
		return out, e
	}
	traceCount := binary.BigEndian.Uint16(p[:2])
	out.Dropped = binary.BigEndian.Uint32(p[2:6])
	if traceCount > TraceCapacity {
		return out, errors.New("invalid trace count")
	}
	counts := []*uint32{&out.Counters.Cycles, &out.Counters.Reads, &out.Counters.Writes, &out.Counters.Claims, &out.Counters.Updates, &out.Counters.Muls, &out.Counters.Adds, &out.Counters.Indirections, &out.Counters.Blackholes, &out.Counters.MaxStack, &out.Counters.Stalls, &out.Counters.Faults}
	for n, v := range counts {
		p, e = s.query(ctx, uint16(16+n))
		if e != nil {
			return out, e
		}
		*v = binary.BigEndian.Uint32(p[6:])
	}
	for n := uint16(0); n < heapSize; n++ {
		p, e = s.query(ctx, 0x1000+n)
		if e != nil {
			return out, e
		}
		out.Heap = append(out.Heap, readWord(p[5:]))
	}
	for n := uint16(0); n < stackSize; n++ {
		p, e = s.query(ctx, 0x2000+n)
		if e != nil {
			return out, e
		}
		out.Stack = append(out.Stack, Frame{Kind: p[0] >> 4, Op: Tag(p[0] & 15), Address: binary.BigEndian.Uint16(p[1:3]), Value: readWord(p[5:])})
	}
	for n := uint16(0); n < traceCount; n++ {
		meta, err := s.query(ctx, 0x3000+2*n)
		if err != nil {
			return out, err
		}
		data, err := s.query(ctx, 0x3001+2*n)
		if err != nil {
			return out, err
		}
		out.Trace = append(out.Trace, Mutation{Cycle: binary.BigEndian.Uint32(meta[:4]), Address: binary.BigEndian.Uint16(meta[4:6]), Old: readWord(data[:5]), New: readWord(data[5:])})
	}
	return out, nil
}
