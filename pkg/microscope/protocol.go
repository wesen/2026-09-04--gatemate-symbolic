package microscope

import (
	"encoding/binary"
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
)

func checksum(p []byte) byte {
	var c byte
	for _, b := range p {
		c ^= b
	}
	return c
}
func EncodeLoad(g Graph) (string, error) {
	a, err := g.Adjacency()
	if err != nil {
		return "", err
	}
	p := []byte{byte(g.Vertices), byte(g.Colors), 0}
	if g.FirstOnly {
		p[2] = 1
	}
	p = append(p, a[:]...)
	p = append(p, checksum(p))
	return "L" + strings.ToUpper(hex.EncodeToString(p)) + "\n", nil
}
func DecodeLoad(line string) (Graph, error) {
	var g Graph
	if len(line) != 26 || line[0] != 'L' || line[25] != '\n' {
		return g, errors.New("invalid load framing")
	}
	p, err := hex.DecodeString(line[1:25])
	if err != nil {
		return g, errors.Wrap(err, "load hexadecimal")
	}
	if checksum(p) != 0 {
		return g, errors.New("load checksum mismatch")
	}
	g.Vertices = int(p[0])
	g.Colors = int(p[1])
	g.FirstOnly = p[2] == 1
	if p[2] > 1 || g.Vertices < 1 || g.Vertices > 8 || g.Colors < 1 || g.Colors > 8 {
		return g, errors.New("invalid load parameters")
	}
	for u := 0; u < 8; u++ {
		for v := 0; v < 8; v++ {
			set := p[3+u]&(1<<v) != 0
			if set && (u == v || u >= g.Vertices || v >= g.Vertices) {
				return g, errors.New("invalid adjacency vertex")
			}
			if set != (p[3+v]&(1<<u) != 0) {
				return g, errors.New("asymmetric adjacency")
			}
			if set && u < v {
				g.Edges = append(g.Edges, [2]int{u, v})
			}
		}
	}
	return g, nil
}
func EncodeEvent(e Event) string {
	p := make([]byte, 33)
	binary.BigEndian.PutUint32(p[0:4], e.Sequence)
	p[4] = e.Kind
	for i, d := range e.Domains {
		p[12-i] = d
	}
	p[13] = e.Propagated
	p[14] = byte(e.ChoiceTop)
	p[15] = byte(e.TrailTop)
	p[16] = byte(e.Base)
	binary.BigEndian.PutUint32(p[17:21], e.Count)
	p[21] = byte(e.Result >> 16)
	p[22] = byte(e.Result >> 8)
	p[23] = byte(e.Result)
	p[24] = e.Fault
	for i := 0; i < 5; i++ {
		p[25+i] = byte(e.ChoiceWord >> uint(8*(4-i)))
	}
	p[30] = byte(e.TrailWord >> 16)
	p[31] = byte(e.TrailWord >> 8)
	p[32] = byte(e.TrailWord)
	p = append(p, checksum(p))
	return "E" + strings.ToUpper(hex.EncodeToString(p)) + "\n"
}
func DecodeEvent(line string) (Event, error) {
	var e Event
	if len(line) != 70 || line[0] != 'E' || line[69] != '\n' {
		return e, errors.New("invalid event framing")
	}
	p, err := hex.DecodeString(line[1:69])
	if err != nil {
		return e, errors.Wrap(err, "event hexadecimal")
	}
	if checksum(p) != 0 {
		return e, errors.New("event checksum mismatch")
	}
	e.Sequence = binary.BigEndian.Uint32(p[:4])
	e.Kind = p[4]
	for i := range e.Domains {
		e.Domains[i] = p[12-i]
	}
	e.Propagated = p[13]
	e.ChoiceTop = int(p[14])
	e.TrailTop = int(p[15])
	e.Base = int(p[16])
	e.Count = binary.BigEndian.Uint32(p[17:21])
	e.Result = uint32(p[21])<<16 | uint32(p[22])<<8 | uint32(p[23])
	e.Fault = p[24]
	for _, b := range p[25:30] {
		e.ChoiceWord = e.ChoiceWord<<8 | uint64(b)
	}
	e.TrailWord = uint32(p[30])<<16 | uint32(p[31])<<8 | uint32(p[32])
	if e.Kind < Create || e.Kind > Fault || e.ChoiceTop > 8 || e.TrailTop > 64 || e.Base > e.TrailTop || e.TrailWord>>20 != 0 {
		return Event{}, errors.New("invalid event fields")
	}
	return e, nil
}
