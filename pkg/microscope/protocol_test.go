package microscope

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestLoadCodec(t *testing.T) {
	g := Graph{3, 3, [][2]int{{0, 1}, {0, 2}, {1, 2}}, true}
	line, err := EncodeLoad(g)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeLoad(line)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := g.Adjacency()
	b, _ := got.Adjacency()
	if a != b || !got.FirstOnly || got.Vertices != 3 || got.Colors != 3 {
		t.Fatal(got)
	}
	for _, bad := range []string{line[:24], strings.Replace(line, "L", "X", 1), line[:1] + "GG" + line[3:], line[:23] + "00\n"} {
		if _, err := DecodeLoad(bad); err == nil {
			t.Fatal("accepted", bad)
		}
	}
	payload, _ := hex.DecodeString(line[1:25])
	payload[3] = 0
	payload[11] = checksum(payload[:11])
	if _, err := DecodeLoad("L" + hex.EncodeToString(payload) + "\n"); err == nil {
		t.Fatal("asymmetry")
	}
}
func TestEventCodecRejectsInvalidRecords(t *testing.T) {
	e := Event{Sequence: 1, Kind: Create, Domains: [8]byte{255, 128, 64, 32, 16, 8, 4, 2}, ChoiceWord: 0x1fc0000000}
	line := EncodeEvent(e)
	if len(line) != 70 {
		t.Fatal(len(line))
	}
	for _, bad := range []string{line[:68], "X" + line[1:], line[:3] + "GG" + line[5:], line[:67] + "FF\n"} {
		if _, err := DecodeEvent(bad); err == nil {
			t.Fatal("accepted", bad)
		}
	}
	e.TrailTop = 65
	if _, err := DecodeEvent(EncodeEvent(e)); err == nil {
		t.Fatal("invalid top")
	}
}
