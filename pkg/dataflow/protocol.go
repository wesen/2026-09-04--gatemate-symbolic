package dataflow

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
	"strings"
)

func EncodeRequest(kind byte, payload []byte) string {
	if len(payload) == 0 {
		return string(kind) + "\n"
	}
	sum := byte(0)
	for _, b := range payload {
		sum ^= b
	}
	data := append(append([]byte{}, payload...), sum)
	return string(kind) + strings.ToUpper(hex.EncodeToString(data)) + "\n"
}
func DecodeRecord(line string, kind byte) ([10]byte, error) {
	var out [10]byte
	if len(line) != 24 || line[0] != kind || line[23] != '\n' {
		return out, errors.New("invalid record framing")
	}
	data, err := hex.DecodeString(line[1:23])
	if err != nil {
		return out, errors.Wrap(err, "decode record hexadecimal")
	}
	sum := byte(0)
	for _, b := range data {
		sum ^= b
	}
	if sum != 0 {
		return out, errors.New("record checksum mismatch")
	}
	copy(out[:], data[:10])
	return out, nil
}
func valueBytes(p []byte) Value { return Value(p[0])<<32 | Value(binary.BigEndian.Uint32(p[1:5])) }

var metricNames = []string{"cycles", "source", "activations", "mul", "alu", "stale", "duplicates", "invalid", "unitBusy", "unitBlocked", "routerBlocked", "outputs", "inputHigh", "readyHigh", "completionHigh", "outputHigh"}

// DecodeSnapshot consumes a complete stopped-machine page set. No result is
// returned until every required page is present and basic bounds are checked.
func DecodeSnapshot(pages map[byte][10]byte) (Snapshot, error) {
	s := Snapshot{Source: "serial", Counters: map[string]uint32{}, Input: []Token{}, Completion: []Token{}, Output: []Token{}}
	page := func(addr byte) ([10]byte, error) {
		p, ok := pages[addr]
		if !ok {
			return p, errors.Errorf("missing snapshot page %d", addr)
		}
		return p, nil
	}
	cap, err := page(0)
	if err != nil {
		return Snapshot{}, err
	}
	if cap[0] != 1 || cap[1] != Contexts || cap[2] != Nodes || cap[3] < 1 || cap[3] > 8 || cap[4] < 1 || cap[4] > 8 || cap[5] < 1 || cap[5] > 8 || cap[6] < 1 || cap[6] > 8 || cap[7] < 1 || cap[7] > 8 {
		return Snapshot{}, errors.New("unsupported device capabilities")
	}
	s.Config = Config{int(cap[4]), int(cap[5]), int(cap[6]), int(cap[3]), int(cap[7])}
	status, err := page(1)
	if err != nil {
		return Snapshot{}, err
	}
	if int(status[5]) > s.Config.InputDepth || status[6] > 28 || int(status[7]) > s.Config.CompletionDepth || int(status[8]) > s.Config.OutputDepth {
		return Snapshot{}, errors.New("invalid snapshot queue counts")
	}
	for c := 0; c < Contexts; c++ {
		s.Epochs[c] = status[3-c]
		s.Closed[c] = status[4]&(1<<(4+c)) != 0
	}
	s.Quiescent = status[9]&128 != 0
	for a := byte(2); a <= 9; a++ {
		p, e := page(a)
		if e != nil {
			return Snapshot{}, e
		}
		s.Counters[metricNames[(a-2)*2]] = binary.BigEndian.Uint32(p[:4])
		s.Counters[metricNames[(a-2)*2+1]] = binary.BigEndian.Uint32(p[4:8])
	}
	for n := byte(0); n < 28; n++ {
		v, e := page(32 + n)
		if e != nil {
			return Snapshot{}, e
		}
		f, e := page(64 + n)
		if e != nil {
			return Snapshot{}, e
		}
		slot := SlotSnapshot{Context: n / 7, Node: n % 7, Valid: f[9] & 3, Issued: f[9]&4 != 0, Pending: f[9]&8 != 0}
		if slot.Valid&1 != 0 {
			slot.Values[0] = valueBytes(v[:5])
		}
		if slot.Valid&2 != 0 {
			slot.Values[1] = valueBytes(v[5:])
		}
		s.Slots = append(s.Slots, slot)
	}
	ip, e := page(10)
	if e != nil {
		return Snapshot{}, e
	}
	iv, e := page(11)
	if e != nil {
		return Snapshot{}, e
	}
	if ip[9]&1 != 0 {
		token, _ := DecodeToken(ip[:])
		phase := int(token.Subtype)
		token.Subtype = 0
		token.Value = 0
		s.Issue = &IssueSnapshot{token, phase, [2]Value{valueBytes(iv[:5]), valueBytes(iv[5:])}}
	}
	rp, e := page(12)
	if e != nil {
		return Snapshot{}, e
	}
	rf, e := page(13)
	if e != nil {
		return Snapshot{}, e
	}
	if rf[9]&1 != 0 {
		t, _ := DecodeToken(rp[:])
		s.Router = &RouterSnapshot{t, (rf[9] >> 1) & 3}
	}
	valid, e := page(24)
	if e != nil {
		return Snapshot{}, e
	}
	s.Mul = make([]*Token, s.Config.MulLatency)
	s.ALU = make([]*Token, 1)
	for n := range s.Mul {
		p, e := page(16 + byte(n))
		if e != nil {
			return Snapshot{}, e
		}
		if valid[9]&(1<<n) != 0 {
			t, _ := DecodeToken(p[:])
			s.Mul[n] = &t
		}
	}
	ap, e := page(25)
	if e != nil {
		return Snapshot{}, e
	}
	if valid[8]&1 != 0 {
		t, _ := DecodeToken(ap[:])
		s.ALU[0] = &t
	}
	for _, q := range []struct {
		base, count byte
		out         *[]Token
	}{{96, status[5], &s.Input}, {112, status[7], &s.Completion}, {128, status[8], &s.Output}} {
		for n := byte(0); n < q.count; n++ {
			p, e := page(q.base + n)
			if e != nil {
				return Snapshot{}, e
			}
			t, _ := DecodeToken(p[:])
			*q.out = append(*q.out, t)
		}
	}
	for c := byte(0); c < Contexts; c++ {
		p, e := page(144 + c)
		if e != nil {
			return Snapshot{}, e
		}
		if status[4]&(1<<c) != 0 {
			t, _ := DecodeToken(p[:])
			s.Errors[c] = &t
		}
	}
	return s, nil
}
