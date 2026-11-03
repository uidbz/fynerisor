package tie

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/tie/runclient"
	"github.com/uidbz/tie/runner/api"
)

// RunnerType is the Risor type of a tie-runner handle: the compute-plane
// entry point (jobs, pipelines, services) next to tie.connect's data plane.
const RunnerType object.Type = "tie.runner"

// Runner wraps one tie-runner URL. Every call goes through runclient, the
// same library the tie-run CLI drives, so scripts and the CLI cannot drift
// from each other.
type Runner struct {
	url string
	// Data-plane endpoint discovered from the runner itself (its
	// GET /api/filehost answer), cached for upload().
	filehost string
	store    string
}

func NewRunner(url string) *Runner { return &Runner{url: url} }

func (r *Runner) Type() object.Type { return RunnerType }
func (r *Runner) Inspect() string   { return "tie.runner(" + r.url + ")" }
func (r *Runner) Interface() any    { return r.url }
func (r *Runner) IsTruthy() bool    { return true }
func (r *Runner) Cost() int         { return 0 }

func (r *Runner) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"type": string(RunnerType), "url": r.url})
}

func (r *Runner) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, fmt.Errorf("tie.runner: unsupported operation %s", opType)
}

func (r *Runner) Equals(other object.Object) bool {
	o, ok := other.(*Runner)
	return ok && o.url == r.url
}

func (r *Runner) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("tie.runner: attribute %q is read-only", name)
}

func (r *Runner) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "url":
		return object.NewString(r.url), true
	case "run":
		return object.NewBuiltin("tie.runner.run", r.Run), true
	case "job":
		return object.NewBuiltin("tie.runner.job", r.GetJob), true
	case "jobs":
		return object.NewBuiltin("tie.runner.jobs", r.Jobs), true
	case "queue":
		return object.NewBuiltin("tie.runner.queue", r.Queue), true
	case "active":
		return object.NewBuiltin("tie.runner.active", r.Active), true
	case "cancel":
		return object.NewBuiltin("tie.runner.cancel", r.CancelJob), true
	case "pipeline":
		return object.NewBuiltin("tie.runner.pipeline", r.SubmitPipeline), true
	case "services":
		return object.NewBuiltin("tie.runner.services", r.Services), true
	case "start_service":
		return object.NewBuiltin("tie.runner.start_service", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			return r.setService(ctx, true, args...)
		}), true
	case "stop_service":
		return object.NewBuiltin("tie.runner.stop_service", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			return r.setService(ctx, false, args...)
		}), true
	case "images":
		return object.NewBuiltin("tie.runner.images", r.Images), true
	case "version":
		return object.NewBuiltin("tie.runner.version", r.Version), true
	case "filehost":
		return object.NewBuiltin("tie.runner.filehost", r.Filehost), true
	case "upload":
		return object.NewBuiltin("tie.runner.upload", r.Upload), true
	}
	return nil, false
}

func (r *Runner) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "url"},
		{Name: "run"},
		{Name: "job"},
		{Name: "jobs"},
		{Name: "queue"},
		{Name: "active"},
		{Name: "cancel"},
		{Name: "pipeline"},
		{Name: "services"},
		{Name: "start_service"},
		{Name: "stop_service"},
		{Name: "images"},
		{Name: "version"},
		{Name: "filehost"},
		{Name: "upload"},
	}
}

// ConnectRunner is tie.runner(url, opts?): a handle on one runner. The
// optional "user" option sets the Basic-Auth user management calls must
// carry (runclient's package default is "tie"; the runner's AuthUser
// config names the matching expectation).
func ConnectRunner(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 {
		return nil, errors.New("tie.runner: url is required")
	}
	url, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.runner: url must be a string (got %s)", args[0].Type())
	}
	if len(url) < 5 {
		return nil, fmt.Errorf("tie.runner: url too short (got %q)", url)
	}
	if len(args) > 2 {
		return nil, fmt.Errorf("tie.runner: expected 1-2 arguments, got %d", len(args))
	}
	if len(args) == 2 {
		optMap, ok := args[1].(*object.Map)
		if !ok {
			return nil, fmt.Errorf("tie.runner: second argument must be a map (got %s)", args[1].Type())
		}
		if v := optMap.GetWithDefault("user", nil); v != nil {
			user, err := object.AsString(v)
			if err != nil {
				return nil, fmt.Errorf("tie.runner: 'user' must be a string (got %s)", v.Type())
			}
			// Deliberately global: the runner client library carries the
			// auth user as a package value (matching the CLI), and a
			// script process talks to one fleet in one dialect.
			runclient.AuthUser = user
		}
	}
	return NewRunner(url), nil
}

// Run submits a job and returns the handle immediately: run(...) then
// job.wait() blocks if the script wants the result, but the job belongs
// to the runner either way (a dropped script does not cancel it).
func (r *Runner) Run(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("tie.runner.run: expected run(image, inputHash) or run(image, inputHash, jobtype), got %d arguments", len(args))
	}
	image, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.runner.run: image must be a string (got %s)", args[0].Type())
	}
	inputHash, err := object.AsString(args[1])
	if err != nil {
		return nil, fmt.Errorf("tie.runner.run: input hash must be a string (got %s)", args[1].Type())
	}
	jobType := "CPUv1"
	if len(args) == 3 {
		jobType, err = object.AsString(args[2])
		if err != nil {
			return nil, fmt.Errorf("tie.runner.run: jobtype must be a string (got %s)", args[2].Type())
		}
	}
	status, err := runclient.Run(r.url, api.JobRequest{
		JobType:   jobType,
		Handler:   image,
		InputHash: inputHash,
	})
	if err != nil {
		return nil, err
	}
	return NewJob(r, status.JobID), nil
}

// GetJob re-adopts a job by id (r.job("...")), e.g. across script runs.
func (r *Runner) GetJob(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.runner.job: expected job id, got %d arguments", len(args))
	}
	id, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.runner.job: id must be a string (got %s)", args[0].Type())
	}
	return NewJob(r, id), nil
}

// Jobs lists recent history: jobs(sinceHours?) — default 24.
func (r *Runner) Jobs(ctx context.Context, args ...object.Object) (object.Object, error) {
	since := 24 * time.Hour
	if len(args) > 0 {
		hours, err := object.AsInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("tie.runner.jobs: sinceHours must be an integer (got %s)", args[0].Type())
		}
		since = time.Duration(hours) * time.Hour
	}
	jobs, errs := runclient.ListJobsAll([]string{r.url}, since, "", "")
	if err, bad := errs[r.url]; bad {
		return nil, err
	}
	out := make([]object.Object, 0, len(jobs))
	for _, j := range jobs {
		m, err := dtoToRisor(j.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return object.NewList(out), nil
}

func (r *Runner) Queue(ctx context.Context, args ...object.Object) (object.Object, error) {
	m, errs := runclient.QueueAll([]string{r.url})
	return r.statusList(ctx, m, errs)
}

func (r *Runner) Active(ctx context.Context, args ...object.Object) (object.Object, error) {
	m, errs := runclient.ActiveAll([]string{r.url})
	return r.statusList(ctx, m, errs)
}

func (r *Runner) statusList(ctx context.Context, m map[string][]*api.JobStatus, errs map[string]error) (object.Object, error) {
	if err, bad := errs[r.url]; bad {
		return nil, err
	}
	out := make([]object.Object, 0, len(m[r.url]))
	for _, j := range m[r.url] {
		mm, err := dtoToRisor(j)
		if err != nil {
			return nil, err
		}
		out = append(out, mm)
	}
	return object.NewList(out), nil
}

// CancelJob cancels a job; the key is the handler image's DeployKey.
func (r *Runner) CancelJob(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("tie.runner.cancel: expected cancel(jobId, deployKey), got %d arguments", len(args))
	}
	id, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.runner.cancel: job id must be a string (got %s)", args[0].Type())
	}
	key, err := object.AsString(args[1])
	if err != nil {
		return nil, fmt.Errorf("tie.runner.cancel: deploy key must be a string (got %s)", args[1].Type())
	}
	if err := runclient.CancelJob(r.url, id, key); err != nil {
		return nil, err
	}
	return object.True, nil
}

// Services lists running tasks (jobs and services share the view).
func (r *Runner) Services(ctx context.Context, args ...object.Object) (object.Object, error) {
	svcs, errs := runclient.ListServicesAll([]string{r.url})
	if err, bad := errs[r.url]; bad {
		return nil, err
	}
	out := make([]object.Object, 0, len(svcs))
	for _, s := range svcs {
		mm, err := dtoToRisor(s.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, mm)
	}
	return object.NewList(out), nil
}

func (r *Runner) setService(ctx context.Context, start bool, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("tie.runner: expected (name, deployKey), got %d arguments", len(args))
	}
	name, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.runner: service name must be a string (got %s)", args[0].Type())
	}
	key, err := object.AsString(args[1])
	if err != nil {
		return nil, fmt.Errorf("tie.runner: deploy key must be a string (got %s)", args[1].Type())
	}
	if err := runclient.SetService(r.url, name, key, start); err != nil {
		return nil, err
	}
	return object.True, nil
}

func (r *Runner) Images(ctx context.Context, args ...object.Object) (object.Object, error) {
	imgs, err := runclient.ListImages(r.url)
	if err != nil {
		return nil, err
	}
	out := make([]object.Object, 0, len(imgs))
	for _, img := range imgs {
		mm, err := dtoToRisor(img)
		if err != nil {
			return nil, err
		}
		out = append(out, mm)
	}
	return object.NewList(out), nil
}

// Version returns the runner's build stamp, capabilities and free capacity.
func (r *Runner) Version(ctx context.Context, args ...object.Object) (object.Object, error) {
	rep, err := runclient.GetVersion(r.url)
	if err != nil {
		return nil, err
	}
	return dtoToRisor(rep)
}

// Filehost reports the data-plane endpoint the runner points jobs at
// ({url, store}); upload() sends input trees there.
func (r *Runner) Filehost(ctx context.Context, args ...object.Object) (object.Object, error) {
	fh, store, err := r.filehostStore(ctx)
	if err != nil {
		return nil, err
	}
	return object.FromGoType(map[string]string{"url": fh, "store": store}), nil
}

// Upload puts a local file or directory on the runner's filehost and
// returns its content address — the inputHash run() wants.
func (r *Runner) Upload(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.runner.upload: expected upload(path), got %d arguments", len(args))
	}
	path, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.runner.upload: path must be a string (got %s)", args[0].Type())
	}
	fh, store, err := r.filehostStore(ctx)
	if err != nil {
		return nil, err
	}
	hash, err := runclient.Upload(fh, path, store)
	if err != nil {
		return nil, err
	}
	return object.NewString(hash), nil
}

func (r *Runner) filehostStore(ctx context.Context) (string, string, error) {
	if r.filehost == "" {
		fh, store, err := runclient.GetFilehostStore(r.url)
		if err != nil {
			return "", "", fmt.Errorf("resolving runner filehost: %w", err)
		}
		r.filehost, r.store = fh, store
	}
	return r.filehost, r.store, nil
}
