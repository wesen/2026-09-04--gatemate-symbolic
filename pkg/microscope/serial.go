package microscope

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.bug.st/serial"
)

type serialPort interface {
	io.ReadWriteCloser
	ResetInputBuffer() error
}
type Serial struct {
	mu                           sync.Mutex
	port                         serialPort
	synchronized, loaded, closed bool
	sequence                     uint32
	// Capture is optional wire evidence; configure it before issuing operations.
	Capture io.Writer
}

var _ Engine = &Serial{}

func NewSerial(device string) (*Serial, error) {
	p, err := serial.Open(device, &serial.Mode{BaudRate: 115200, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		return nil, errors.Wrap(err, "open FPGA serial device")
	}
	if err := p.SetReadTimeout(50 * time.Millisecond); err != nil {
		_ = p.Close()
		return nil, errors.Wrap(err, "set serial timeout")
	}
	return &Serial{port: p}, nil
}
func (s *Serial) resynchronize(ctx context.Context) error {
	if s.closed {
		return errors.New("serial device is closed")
	}
	if s.synchronized {
		return nil
	}
	// Let the FPGA's 200 ms incomplete-command timeout and any response drain.
	timer := time.NewTimer(250 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	if err := s.port.ResetInputBuffer(); err != nil {
		return errors.Wrap(err, "drain serial input")
	}
	return nil
}
func (s *Serial) exchange(ctx context.Context, command string) (response string, err error) {
	if s.closed {
		return "", errors.New("serial device is closed")
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
	bytes := []byte(command)
	for len(bytes) > 0 {
		if err = ctx.Err(); err != nil {
			return "", err
		}
		n, e := s.port.Write(bytes)
		if e != nil {
			return "", errors.Wrap(e, "write serial command")
		}
		if n <= 0 {
			return "", io.ErrShortWrite
		}
		bytes = bytes[n:]
	}
	var line []byte
	buf := make([]byte, 80)
	for {
		if err = ctx.Err(); err != nil {
			return "", errors.Wrap(err, "serial response timeout or cancellation")
		}
		n, e := s.port.Read(buf)
		if e != nil {
			return "", errors.Wrap(e, "read serial response")
		}
		line = append(line, buf[:n]...)
		if len(line) > 128 {
			return "", errors.New("serial response exceeds framing bound")
		}
		if i := strings.IndexByte(string(line), '\n'); i >= 0 {
			if i != len(line)-1 {
				return "", errors.New("unexpected bytes after response")
			}
			response = string(line)
			if s.Capture != nil {
				_, _ = fmt.Fprintf(s.Capture, "< %s", response)
			}
			if strings.HasPrefix(response, "!") {
				return "", errors.Errorf("FPGA rejected command: %s", strings.TrimSpace(response))
			}
			return response, nil
		}
	}
}
func (s *Serial) Load(ctx context.Context, g Graph) error {
	command, err := EncodeLoad(g)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.resynchronize(ctx); err != nil {
		return err
	}
	response, err := s.exchange(ctx, command)
	if err != nil {
		return err
	}
	if response != "A\n" {
		s.synchronized = false
		return errors.New("expected graph acknowledgement")
	}
	s.loaded = true
	s.synchronized = true
	s.sequence = 0
	return nil
}
func (s *Serial) Reset(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.loaded {
		return errors.New("load a graph first")
	}
	if err := s.resynchronize(ctx); err != nil {
		return err
	}
	response, err := s.exchange(ctx, "R\n")
	if err != nil {
		return err
	}
	if response != "A\n" {
		s.synchronized = false
		return errors.New("expected reset acknowledgement")
	}
	s.sequence = 0
	s.synchronized = true
	return nil
}
func (s *Serial) Step(ctx context.Context) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.loaded {
		return Event{}, errors.New("load a graph first")
	}
	if !s.synchronized {
		return Event{}, errors.New("serial link is unsynchronized; reset or reload explicitly")
	}
	response, err := s.exchange(ctx, "S\n")
	if err != nil {
		return Event{}, err
	}
	if response == "Z\n" {
		return Event{}, ErrTerminal
	}
	e, err := DecodeEvent(response)
	if err != nil {
		s.synchronized = false
		return Event{}, err
	}
	if e.Sequence != s.sequence+1 {
		s.synchronized = false
		return Event{}, errors.New("serial event sequence gap")
	}
	s.sequence = e.Sequence
	return e, nil
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
