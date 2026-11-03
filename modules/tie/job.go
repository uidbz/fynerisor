package tie

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/tie/runclient"
	"github.com/uidbz/tie/runner/api"
)

// JobType is the Risor type of a submitted job's handle.
const JobType object.Type = "tie.job"

// Job is a handle on one runner job. It is a handle, not a snapshot: the
// runner owns the record, and every read re-fetches it (so a handle held
// across a script pause always sees live state).
type Job struct {
	runner *Runner
	id     string
}

func NewJob(r *Runner, id string) *Job { return &Job{runner: r, id: id} }

func (j *Job) Type() object.Type { return JobType }
func (j *Job) Inspect() string   { return "tie.job(" + j.id + ")" }
func (j *Job) Interface() any    { return j.id }
func (j *Job) IsTruthy() bool    { return true }
func (j *Job) Cost() int         { return 0 }

func (j *Job) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"type": string(JobType), "jobid": j.id})
}

func (j *Job) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, fmt.Errorf("tie.job: unsupported operation %s", opType)
}

func (j *Job) Equals(other object.Object) bool {
	o, ok := other.(*Job)
	return ok && o.id == j.id && o.runner.url == j.runner.url
}

func (j *Job) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("tie.job: attribute %q is read-only", name)
}

func (j *Job) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "id":
		return object.NewString(j.id), true
	case "status":
		return object.NewBuiltin("tie.job.status", j.Status), true
	case "status_str":
		return object.NewBuiltin("tie.job.status_str", j.scalar(func(s api.JobStatus) any { return s.Status })), true
	case "done":
		return object.NewBuiltin("tie.job.done", j.scalar(func(s api.JobStatus) any { return s.Done })), true
	case "percent":
		return object.NewBuiltin("tie.job.percent", j.scalar(func(s api.JobStatus) any { return s.Percent })), true
	case "output_hash":
		return object.NewBuiltin("tie.job.output_hash", j.scalar(func(s api.JobStatus) any { return s.OutputHash })), true
	case "error_msg":
		return object.NewBuiltin("tie.job.error_msg", j.scalar(func(s api.JobStatus) any { return s.ErrorMsg })), true
	case "log":
		return object.NewBuiltin("tie.job.log", j.scalar(func(s api.JobStatus) any { return s.Log })), true
	case "wait":
		return object.NewBuiltin("tie.job.wait", j.Wait), true
	case "cancel":
		return object.NewBuiltin("tie.job.cancel", j.Cancel), true
	}
	return nil, false
}

func (j *Job) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "id"},
		{Name: "status"},
		{Name: "status_str"},
		{Name: "done"},
		{Name: "percent"},
		{Name: "output_hash"},
		{Name: "error_msg"},
		{Name: "log"},
		{Name: "wait"},
		{Name: "cancel"},
	}
}

// dtoToRisor converts a runner DTO into an indexable Risor map via its
// JSON form: scripts see exactly the field spellings docs/runner-api.md
// documents, and the default registry's opaque go_struct never leaks.
func dtoToRisor(v any) (object.Object, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	objs, err := object.AsObjects(m)
	if err != nil {
		return nil, err
	}
	return object.NewMap(objs), nil
}

func (j *Job) fetch(ctx context.Context) (api.JobStatus, error) {
	return runclient.GetStatus(j.runner.url, j.id)
}

// Status returns the full wire record (the same field names
// docs/runner-api.md documents), refreshing it.
func (j *Job) Status(ctx context.Context, args ...object.Object) (object.Object, error) {
	s, err := j.fetch(ctx)
	if err != nil {
		return nil, err
	}
	return dtoToRisor(s)
}

func (j *Job) scalar(pick func(api.JobStatus) any) func(ctx context.Context, args ...object.Object) (object.Object, error) {
	return func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("tie.job: this attribute takes no arguments")
		}
		s, err := j.fetch(ctx)
		if err != nil {
			return nil, err
		}
		return object.FromGoType(pick(s)), nil
	}
}

// Wait blocks until the job reaches a terminal state and returns the
// final record. wait() waits indefinitely (the runner keeps the job no
// matter what the script does); wait(timeoutSeconds) gives up with an
// error after the bound, leaving the job running.
func (j *Job) Wait(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("tie.job.wait: expected wait() or wait(timeoutSeconds), got %d arguments", len(args))
	}
	deadline := time.Time{}
	if len(args) == 1 {
		secs, err := object.AsInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("tie.job.wait: timeout must be an integer number of seconds (got %s)", args[0].Type())
		}
		deadline = time.Now().Add(time.Duration(secs) * time.Second)
	}
	for {
		s, err := j.fetch(ctx)
		if err != nil {
			return nil, err
		}
		if s.Done {
			return dtoToRisor(s)
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return nil, fmt.Errorf("tie.job.wait: job %s still %s after timeout (it keeps running on the runner)", j.id, s.Status)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("tie.job.wait: watching cancelled; job %s keeps running on the runner", j.id)
		case <-time.After(300 * time.Millisecond):
		}
	}
}

// Cancel asks the runner to stop the job; the key is the handler image's
// DeployKey.
func (j *Job) Cancel(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.job.cancel: expected cancel(deployKey), got %d arguments", len(args))
	}
	key, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.job.cancel: deploy key must be a string (got %s)", args[0].Type())
	}
	if err := runclient.CancelJob(j.runner.url, j.id, key); err != nil {
		return nil, err
	}
	return object.True, nil
}
