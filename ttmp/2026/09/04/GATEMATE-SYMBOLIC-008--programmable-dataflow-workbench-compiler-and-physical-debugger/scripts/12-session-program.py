from pathlib import Path
p=Path('internal/dataflowide/session.go');s=p.read_text().replace('type Frame struct {','type Frame struct {\n Program *df.Program `json:"program"`').replace('type Session struct {','type Session struct {\n program *df.Program')
s=s.replace('state.Current.Snapshot = state.Current.Snapshot.Clone()','state.Current.Snapshot = state.Current.Snapshot.Clone()\n state.Current.Program=state.Current.Program.Clone()')
s=s.replace('f.Snapshot = f.Snapshot.Clone()','f.Snapshot = f.Snapshot.Clone()\n f.Program=f.Program.Clone()')
s=s.replace('Frame{FrameInfo{s.nextID, s.generation, label, snap.Counters["cycles"], time.Now().UTC().Format(time.RFC3339Nano)}, snap}', 'Frame{Program:s.program.Clone(),FrameInfo:FrameInfo{s.nextID, s.generation, label, snap.Counters["cycles"], time.Now().UTC().Format(time.RFC3339Nano)}, Snapshot:snap}')
s=s.replace('if o.Kind == "reset" {','if o.Kind=="load" {s.program=nil}\n if o.Kind == "reset" {\n s.program=nil',1)
p.write_text(s)
p=Path('internal/dataflowide/http.go');s=p.read_text().replace('mux := http.NewServeMux()', 'mux := http.NewServeMux()\n programRoutes(mux,s,p)');p.write_text(s)
p=Path('internal/dataflowide/projects.go');s=p.read_text()
s=s.replace('func (p *Projects) List() ([]string, error) {','func (p *Projects) List() ([]string,error) {return p.list(".json")}\nfunc (p *Projects) list(extension string) ([]string, error) {')
s=s.replace('strings.TrimSuffix(e.Name(), ".json")','strings.TrimSuffix(e.Name(), extension)').replace('strings.HasSuffix(e.Name(), ".json")','strings.HasSuffix(e.Name(), extension)')
s=s.replace('func (p *Projects) Read(id string) (string, error) {','func (p *Projects) Read(id string) (string,error) {return p.read(id,".json")}\nfunc (p *Projects) read(id,extension string) (string, error) {').replace('p.root.Open(id + ".json")','p.root.Open(id + extension)')
a=s.index('func (p *Projects) Write(');b=s.index('\tvar nonce',a)
s=s[:a]+'''func (p *Projects) Write(id,source string) error {
 _,diagnostics:=ParseScenario(source);if len(diagnostics)!=0{return errors.New("validate and fix the scenario before saving")}
 return p.write(id,source,".json")
}
func (p *Projects) write(id,source,extension string) error {
 if !projectID.MatchString(id){return errors.New("project id must use lowercase letters, digits, and hyphens")}
'''+s[b:];s=s.replace('id+".json"','id+extension');p.write_text(s)
