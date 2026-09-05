package dataflow

import (
	"context"
	"encoding/binary"
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
		return nil, errors.Wrap(err, "open dataflow UART")
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
		if err != nil && !errors.Is(err, ErrFull) && !errors.Is(err, ErrBlocked) {
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
			return "", errors.Wrap(err, "write dataflow request")
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
			return "", errors.Wrap(err, "read dataflow response")
		}
		n, e := s.port.Read(buffer)
		if e != nil {
			return "", errors.Wrap(e, "read dataflow UART")
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
			switch line {
			case "!04\n":
				return "", ErrFull
			case "!05\n":
				return "", ErrBlocked
			}
			if strings.HasPrefix(line, "!") {
				return "", errors.Errorf("device rejected request: %s", strings.TrimSpace(line))
			}
			return line, nil
		}
	}
}
func (s *Serial) Execute(ctx context.Context, o Operation) (*Token, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("serial device closed")
	}
	if o.Config != nil {
		return nil, errors.New("hardware configuration is fixed by its bitstream")
	}
	if o.Kind == "reset" && !s.synchronized {
		timer := time.NewTimer(250 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
		}
		if err := s.port.ResetInputBuffer(); err != nil {
			return nil, errors.Wrap(err, "drain UART before explicit reset")
		}
	} else if !s.synchronized {
		return nil, errors.New("UART state uncertain; explicit reset required")
	}
	if o.Kind == "load" {
		for n := byte(0); n < o.Graph.Count; n++ {
			bytes := o.Graph.Bytes(n)
			line, err := s.exchange(ctx, EncodeRequest('W', bytes[:]))
			if err != nil {
				return nil, err
			}
			if line != "A\n" {
				s.synchronized = false
				return nil, errors.New("expected descriptor acknowledgement")
			}
		}
		line, err := s.exchange(ctx, EncodeRequest('G', []byte{o.Graph.Count}))
		if err != nil {
			return nil, err
		}
		if line != "A\n" {
			s.synchronized = false
			return nil, errors.New("expected graph activation acknowledgement")
		}
		return nil, nil
	}
	var command string
	switch o.Kind {
	case "reset":
		command = "R\n"
	case "inject":
		p, _ := o.Token.Bytes()
		command = EncodeRequest('I', p[:])
	case "tick":
		var p [4]byte
		binary.BigEndian.PutUint32(p[:], o.Ticks)
		command = EncodeRequest('T', p[:])
	case "cancel":
		command = EncodeRequest('C', []byte{o.Context})
	case "poll":
		command = "P\n"
	}
	line, err := s.exchange(ctx, command)
	if err != nil {
		return nil, err
	}
	if o.Kind == "poll" {
		if line == "N\n" {
			return nil, nil
		}
		p, err := DecodeRecord(line, 'O')
		if err != nil {
			s.synchronized = false
			return nil, err
		}
		t, err := DecodeToken(p[:])
		return &t, err
	}
	if line != "A\n" {
		s.synchronized = false
		return nil, errors.New("expected operation acknowledgement")
	}
	s.synchronized = true
	return nil, nil
}
func (s *Serial) Snapshot(ctx context.Context) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.synchronized {
		return Snapshot{}, errors.New("UART state uncertain; explicit reset required")
	}
	pages := map[byte][10]byte{}
	// Read a fixed bounded set. Hardware remains paused throughout and no other
	// operation can interleave through this Serial instance.
	ranges := [][2]byte{{0, 13}, {16, 25}, {32, 59}, {64, 91}, {96, 103}, {112, 119}, {128, 135}, {144, 155}}
	for _, r := range ranges {
		for addr := r[0]; addr <= r[1]; addr++ {
			line, err := s.exchange(ctx, EncodeRequest('Q', []byte{addr}))
			if err != nil {
				return Snapshot{}, err
			}
			p, err := DecodeRecord(line, 'S')
			if err != nil {
				s.synchronized = false
				return Snapshot{}, err
			}
			pages[addr] = p
		}
	}
	snapshot, err := DecodeSnapshot(pages)
	if err != nil {
		s.synchronized = false
	}
	return snapshot, err
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
