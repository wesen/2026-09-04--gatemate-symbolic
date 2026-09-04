package microscope

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

type fakePort struct {
	input, command bytes.Buffer
	model          *Model
	corrupt        bool
	silent         bool
	closed         bool
}

func (p *fakePort) Write(data []byte) (int, error) {
	n := len(data)
	if n > 3 {
		n = 3
	}
	p.command.Write(data[:n])
	line := p.command.String()
	if strings.HasSuffix(line, "\n") {
		p.command.Reset()
		var response string
		switch line[0] {
		case 'L':
			g, err := DecodeLoad(line)
			if err != nil {
				return 0, err
			}
			_ = p.model.Load(context.Background(), g)
			response = "A\n"
		case 'R':
			_ = p.model.Reset(context.Background())
			response = "A\n"
		case 'S':
			e, err := p.model.Step(context.Background())
			if err != nil {
				response = "Z\n"
			} else {
				response = EncodeEvent(e)
			}
			if p.corrupt {
				response = response[:67] + "FF\n"
			}
		}
		if !p.silent {
			p.input.WriteString(response)
		}
	}
	return n, nil
}
func (p *fakePort) Read(data []byte) (int, error) {
	if p.input.Len() == 0 {
		time.Sleep(time.Millisecond)
		return 0, nil
	}
	if len(data) > 3 {
		data = data[:3]
	}
	return p.input.Read(data)
}
func (p *fakePort) ResetInputBuffer() error { p.input.Reset(); return nil }
func (p *fakePort) Close() error            { p.closed = true; return nil }

var _ serialPort = &fakePort{}

func TestSerialFragmentsAndExplicitRecovery(t *testing.T) {
	p := &fakePort{model: NewModel()}
	s := &Serial{port: p, synchronized: true}
	ctx := context.Background()
	g := Graph{Vertices: 2, Colors: 2}
	if err := s.Load(ctx, g); err != nil {
		t.Fatal(err)
	}
	e, err := s.Step(ctx)
	if err != nil || e.Sequence != 1 {
		t.Fatal(e, err)
	}
	p.corrupt = true
	if _, err := s.Step(ctx); err == nil {
		t.Fatal("accepted corrupted response")
	}
	if _, err := s.Step(ctx); err == nil {
		t.Fatal("silently retried unsynchronized link")
	}
	p.corrupt = false
	if err := s.Reset(ctx); err != nil {
		t.Fatal(err)
	}
	e, err = s.Step(ctx)
	if err != nil || e.Sequence != 1 {
		t.Fatal(e, err)
	}
	if err := s.Close(); err != nil || !p.closed {
		t.Fatal(err)
	}
	if _, err := s.Step(ctx); err == nil {
		t.Fatal("closed device")
	}
}
func TestSerialCancelledReadAndInvalidLoad(t *testing.T) {
	p := &fakePort{model: NewModel()}
	s := &Serial{port: p, synchronized: true}
	ctx := context.Background()
	if err := s.Load(ctx, Graph{}); err == nil {
		t.Fatal("invalid graph")
	}
	if p.command.Len() != 0 {
		t.Fatal("invalid graph touched device")
	}
	_ = s.Load(ctx, Graph{Vertices: 2, Colors: 2})
	p.silent = true
	limited, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	if _, err := s.Step(limited); err == nil {
		t.Fatal("unbounded read")
	}
	if s.synchronized {
		t.Fatal("timeout did not poison link")
	}
}
