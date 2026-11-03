package tie

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/uidbz/tie/io/compose"
	"github.com/uidbz/tie/metadata"
	tieapi "github.com/uidbz/tie/runner/api"
	"github.com/uidbz/tie/runner/dispatch"
	"github.com/uidbz/tie/runner/executor"
	"github.com/uidbz/tie/runner/fleet"
	"github.com/uidbz/tie/runner/images"
	"github.com/uidbz/tie/runner/pipeline"
	"github.com/uidbz/tie/runner/proxy"
	"github.com/uidbz/tie/runner/scheduler"
	"github.com/uidbz/tie/runner/sec"
	"github.com/uidbz/tie/runner/server"
	"github.com/uidbz/tie/runner/service"
	"github.com/uidbz/tie/runner/store"
)

// The script surface is only as good as an end-to-end script run, so these
// tests evaluate real Risor source against a real assembled tie-runner
// (exported API, fake store and fake container runtime): the same wire the
// CLI speaks, through the bindings, through the interpreter. The scripts
// end with a summary map; the assertions live in Go.

// --- fake filehost: GET /{hash} serves, PUT stores under the computed
// address and echoes it back (the shape compose/putlib expect) ----------

func fakeFilehost(t *testing.T, blobs map[string]string) string {
	t.Helper()
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			hash := strings.TrimPrefix(r.URL.Path, "/")
			mu.Lock()
			body, ok := blobs[hash]
			mu.Unlock()
			if !ok {
				http.Error(w, "no such blob", http.StatusNotFound)
				return
			}
			io.WriteString(w, body)
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			addr, err := metadata.HashReader(strings.NewReader(string(body)))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			mu.Lock()
			blobs[addr] = string(body)
			mu.Unlock()
			io.WriteString(w, addr)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func uploadTree(t *testing.T, filehost, name, content string) string {
	t.Helper()
	b := compose.NewTreeBuilder()
	addr, err := metadata.HashReader(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Add(name, metadata.DirEntry{Hash: addr, Filename: name, Size: len(content)}); err != nil {
		t.Fatal(err)
	}
	tree, err := b.Upload(filehost, "")
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

// --- fake container runtime: reports a fixed output tree through the
// hooks sink, like the real Docker runtime does --------------------------

type fakeLocal struct {
	mu      sync.Mutex
	stopped map[string]chan struct{}
	result  string
}

func (f *fakeLocal) Run(ctx context.Context, job *tieapi.JobStatus, opts executor.RunOptions) executor.ExitInfo {
	name := executor.ContainerName(job.Handler)
	ch := make(chan struct{})
	f.mu.Lock()
	if f.stopped == nil {
		f.stopped = map[string]chan struct{}{}
	}
	f.stopped[name] = ch
	f.mu.Unlock()
	defer func() {
		f.mu.Lock()
		delete(f.stopped, name)
		f.mu.Unlock()
	}()
	if opts.Hooks.Result != nil {
		opts.Hooks.Result(f.result)
	}
	return executor.ExitInfo{}
}

func (f *fakeLocal) Stop(containerName string) error {
	f.mu.Lock()
	ch := f.stopped[containerName]
	f.mu.Unlock()
	if ch == nil {
		return fmt.Errorf("no container %s", containerName)
	}
	select {
	case <-ch:
	default:
		close(ch)
	}
	return nil
}

// --- assembled runner (the same shape cmd/tie-runner wires) --------------

type runnerEnv struct {
	url        string
	filehost   string
	resultHash string
}

func newRunnerEnv(t *testing.T) *runnerEnv {
	t.Helper()

	blobs := map[string]string{}
	fhURL := fakeFilehost(t, blobs)
	resultHash := uploadTree(t, fhURL, "out.dat", "job output")

	st := store.NewFake()
	local := &fakeLocal{result: resultHash}
	imgReg := images.New(func([]images.Image) error { return nil },
		images.Image{Name: "tool:v2", DeployKey: "dk"})
	svc := service.New(service.Config{
		Images: imgReg, Bin: "docker-does-not-exist",
		ExtDir: t.TempDir(), Ports: service.NewPortPool(3000, 3100),
	})

	var sched *scheduler.Manager
	var srv *server.Server
	pipe := pipeline.New(pipeline.Config{
		Store:    st,
		Filehost: fhURL,
		SubmitJob: func(req tieapi.JobRequest, dm string) (*tieapi.JobStatus, error) {
			return srv.SubmitJob(req, dm)
		},
		CancelJob: func(jobID string) error { return sched.Cancel(jobID) },
		ValidateStep: func(req tieapi.JobRequest) error {
			return srv.ValidateStep(req)
		},
	})

	sched = scheduler.New(scheduler.Config{
		Local: local, Store: st,
		TotalCPU: 8, TotalRAMMB: 65536, TotalGPU: 2,
		Filehost:   fhURL,
		OnFinished: pipe.JobFinished,
	})

	srv = server.New(server.Deps{
		Store: st, Sched: sched, Pipe: pipe, Svc: svc,
		Images: imgReg, Auth: &sec.Auth{},
		Fleet: fleet.Load(filepath.Join(t.TempDir(), "fleet-registry.json")),
		Dispatch: &dispatch.Router{
			Capabilities: func() []string { return []string{"CPUv1"} },
			ResolveSpec:  sched.SpecFor,
			CanFit:       sched.CanFit,
		},
		Proxy:        proxy.New(proxy.Config{ExternalPort: svc.RunningPort}),
		Filehost:     func() string { return fhURL },
		StoreName:    func() string { return "" },
		Capabilities: func() []string { return []string{"CPUv1"} },
		Version:      "v-test",
	})

	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return &runnerEnv{url: ts.URL, filehost: fhURL, resultHash: resultHash}
}

func evalScript(t *testing.T, src string) (object.Object, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// Same shape fynerisor core assembles: standard builtins with the
	// tie module registered alongside them.
	env := risor.Builtins()
	env["tie"] = Module()
	res, err := risor.Eval(ctx, src, risor.WithEnv(env))
	if err != nil {
		return nil, err
	}
	obj, ok := res.(object.Object)
	if !ok {
		// The script's final map literal comes back as a plain Go map;
		// put it back through the registry so the helpers can index it.
		conv := object.FromGoType(res)
		if conv == nil {
			return nil, fmt.Errorf("script result is not a risor object: %T", res)
		}
		obj = conv
	}
	return obj, nil
}

func mapField(t *testing.T, obj object.Object, key string) object.Object {
	t.Helper()
	m, ok := obj.(*object.Map)
	if !ok {
		t.Fatalf("script returned %s, want a map", obj.Type())
	}
	v := m.Get(key)
	if v == nil || v.Type() == object.NIL {
		t.Fatalf("script result missing %q: %s", key, m.Inspect())
	}
	return v
}

func wantBool(t *testing.T, obj object.Object, key string, want bool) {
	t.Helper()
	got := mapField(t, obj, key)
	b, err := object.AsBool(got)
	if err != nil || b != want {
		t.Errorf("%s = %s, want %v", key, got.Inspect(), want)
	}
}

func wantString(t *testing.T, obj object.Object, key, want string) {
	t.Helper()
	got := mapField(t, obj, key)
	s, err := object.AsString(got)
	if err != nil || s != want {
		t.Errorf("%s = %s, want %q", key, got.Inspect(), want)
	}
}

func wantInt(t *testing.T, obj object.Object, key string, want int64) {
	t.Helper()
	got := mapField(t, obj, key)
	n, err := object.AsInt(got)
	if err != nil || n != want {
		t.Errorf("%s = %s, want %d", key, got.Inspect(), want)
	}
}

// TestScriptRunAndWait drives the whole job lifecycle from Risor source:
// handle, submit, wait, scalars, history, version.
func TestScriptRunAndWait(t *testing.T) {
	env := newRunnerEnv(t)
	seed := strings.Repeat("ab", 32) // 64 hex chars: the shape validate enforces

	res, err := evalScript(t, `
let r = tie.runner("`+env.url+`")
let v = r.version()
let j = r.run("tool:v2", "`+seed+`", "CPUv1")
let final = j.wait(30)
{
  "version": v["Version"],
  "status": final["Status"],
  "hash": final["OutputHash"],
  "done": j.done(),
  "njobs": len(r.jobs(24)),
  "idmatches": j.id == final["JobID"],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	wantString(t, res, "version", "v-test")
	wantString(t, res, "status", tieapi.StatusDone)
	wantString(t, res, "hash", env.resultHash)
	wantInt(t, res, "njobs", 1)
	wantBool(t, res, "done", true)
	wantBool(t, res, "idmatches", true) // JobID round-trips through the handle
}

// TestScriptPipeline drives a two-step composed pipeline from Risor:
// seeds, step inputs referencing predecessors, wait, steps inspection,
// plus the images list.
func TestScriptPipeline(t *testing.T) {
	env := newRunnerEnv(t)
	seed := uploadTree(t, env.filehost, "in.dat", "seed content")

	res, err := evalScript(t, `
let r = tie.runner("`+env.url+`")
let p = r.pipeline({
  "name": "chain",
  "seeds": { "raw": "`+seed+`" },
  "steps": [
    { "name": "one", "image": "tool:v2", "jobtype": "CPUv1", "inputs": [{ "from": "raw" }] },
    { "name": "two", "image": "tool:v2", "jobtype": "CPUv1", "inputs": [{ "from": "one" }] },
  ],
})
let fin = p.wait()
let steps = p.steps()
let doneCount = len(steps.filter((s) => s["status"] == "done" && s["output_hash"] == "`+env.resultHash+`"))

{
  "status": fin["status"],
  "done": p.done(),
  "doneSteps": doneCount,
  "images": len(r.images()),
}
`)
	if err != nil {
		t.Fatal(err)
	}
	wantString(t, res, "status", tieapi.StatusDone)
	wantInt(t, res, "doneSteps", 2)
	wantInt(t, res, "images", 1)
	wantBool(t, res, "done", true)
}

// TestScriptErrorsSurface asserts that runner-side failures cross the
// binding as Risor errors (wrong job type, unknown job, bad pipeline
// request): scripts get honest exceptions, not nil maps.
func TestScriptErrorsSurface(t *testing.T) {
	env := newRunnerEnv(t)
	seed := strings.Repeat("cd", 32)

	for i, src := range []string{
		// unknown job type: runner rejects at submit
		`let r = tie.runner("` + env.url + `")
		 let j = r.run("tool:v2", "` + seed + `", "NoSuchType")
		 j.id()`,
		// unknown job id on re-adoption then status
		`let r = tie.runner("` + env.url + `")
		 let j = r.job("no-such-job")
		 j.status_str()`,
		// malformed pipeline request caught client-side
		`let r = tie.runner("` + env.url + `")
		 let p = r.pipeline({ "steps": [{ "name": "one" }] })
		 p.id()`,
		// wrong deploy key for cancel must error, not silently pass
		`let r = tie.runner("` + env.url + `")
		 let j = r.run("tool:v2", "` + seed + `", "CPUv1")
		 j.wait(30)
		 j.cancel("wrong-key")`,
	} {
		if _, err := evalScript(t, src); err == nil {
			t.Errorf("case %d: expected a Risor error, script succeeded", i)
		}
	}
}
