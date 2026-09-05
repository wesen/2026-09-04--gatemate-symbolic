package dataflowide

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

var projectID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

type Projects struct{ root *os.Root }

func NewProjects(path string) (*Projects, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, errors.Wrap(err, "create project directory")
	}
	r, err := os.OpenRoot(path)
	if err != nil {
		return nil, errors.Wrap(err, "open project root")
	}
	return &Projects{r}, nil
}
func (p *Projects) Close() error { return p.root.Close() }
func (p *Projects) List() ([]string, error) {
	f, err := p.root.Open(".")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entries, err := f.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, e := range entries {
		id := strings.TrimSuffix(e.Name(), ".json")
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") && projectID.MatchString(id) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids, nil
}
func (p *Projects) Read(id string) (string, error) {
	if !projectID.MatchString(id) {
		return "", errors.New("invalid project id")
	}
	f, err := p.root.Open(id + ".json")
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, MaxSourceBytes+1))
	if err != nil {
		return "", err
	}
	if len(b) > MaxSourceBytes {
		return "", errors.New("project too large")
	}
	return string(b), nil
}
func (p *Projects) Write(id, source string) error {
	if !projectID.MatchString(id) {
		return errors.New("project id must use lowercase letters, digits, and hyphens")
	}
	_, diagnostics := ParseScenario(source)
	if len(diagnostics) != 0 {
		return errors.New("validate and fix the scenario before saving")
	}
	var nonce [12]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	temporary := ".save-" + hex.EncodeToString(nonce[:])
	f, err := p.root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer p.root.Remove(temporary)
	if _, err = f.WriteString(source); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return p.root.Rename(temporary, id+".json")
}
