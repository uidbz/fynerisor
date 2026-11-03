package tie

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/tie/io/compose"
	"github.com/uidbz/tie/runclient"
	"github.com/uidbz/tie/runner/api"
)

// PipelineType is the Risor type of a submitted pipeline run.
const PipelineType object.Type = "tie.pipeline"

// Pipeline is a handle on one runner pipeline: the runner sequences the
// steps, the handle only watches and controls.
type Pipeline struct {
	runner *Runner
	id     string
}

func NewPipeline(r *Runner, id string) *Pipeline { return &Pipeline{runner: r, id: id} }

func (p *Pipeline) Type() object.Type { return PipelineType }
func (p *Pipeline) Inspect() string   { return "tie.pipeline(" + p.id + ")" }
func (p *Pipeline) Interface() any    { return p.id }
func (p *Pipeline) IsTruthy() bool    { return true }
func (p *Pipeline) Cost() int         { return 0 }

func (p *Pipeline) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"type": string(PipelineType), "pipelineid": p.id})
}

func (p *Pipeline) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, fmt.Errorf("tie.pipeline: unsupported operation %s", opType)
}

func (p *Pipeline) Equals(other object.Object) bool {
	o, ok := other.(*Pipeline)
	return ok && o.id == p.id && o.runner.url == p.runner.url
}

func (p *Pipeline) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("tie.pipeline: attribute %q is read-only", name)
}

func (p *Pipeline) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "id":
		return object.NewString(p.id), true
	case "status":
		return object.NewBuiltin("tie.pipeline.status", p.Status), true
	case "steps":
		return object.NewBuiltin("tie.pipeline.steps", p.Steps), true
	case "done":
		return object.NewBuiltin("tie.pipeline.done", p.Done), true
	case "wait":
		return object.NewBuiltin("tie.pipeline.wait", p.Wait), true
	case "cancel":
		return object.NewBuiltin("tie.pipeline.cancel", p.Cancel), true
	case "resume":
		return object.NewBuiltin("tie.pipeline.resume", p.Resume), true
	}
	return nil, false
}

func (p *Pipeline) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "id"},
		{Name: "status"},
		{Name: "steps"},
		{Name: "done"},
		{Name: "wait"},
		{Name: "cancel"},
		{Name: "resume"},
	}
}

func (p *Pipeline) fetch(ctx context.Context) (*api.Pipeline, error) {
	pp, err := runclient.GetPipeline(p.runner.url, p.id)
	if err != nil {
		return nil, err
	}
	return &pp, nil
}

// Status returns the full wire record (status plus per-step state).
func (p *Pipeline) Status(ctx context.Context, args ...object.Object) (object.Object, error) {
	pp, err := p.fetch(ctx)
	if err != nil {
		return nil, err
	}
	return pipelineToRisor(pp), nil
}

// Steps lists the step records (name/status/jobid/output_hash/error).
func (p *Pipeline) Steps(ctx context.Context, args ...object.Object) (object.Object, error) {
	pp, err := p.fetch(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]object.Object, 0, len(pp.Steps))
	for _, s := range pp.Steps {
		out = append(out, stepToRisor(s))
	}
	return object.NewList(out), nil
}

func (p *Pipeline) Done(ctx context.Context, args ...object.Object) (object.Object, error) {
	pp, err := p.fetch(ctx)
	if err != nil {
		return nil, err
	}
	return object.FromGoType(runclient.PipelineDone(pp)), nil
}

// Wait blocks until the run reaches a terminal state and returns the
// final record. The run does not depend on the wait: giving up (context
// cancel, script end) leaves it running on the runner.
func (p *Pipeline) Wait(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("tie.pipeline.wait: takes no arguments")
	}
	pp, err := runclient.WaitPipelineCtx(ctx, p.runner.url, p.id, 300_000_000, nil) // 300ms
	if err != nil {
		return nil, err
	}
	return pipelineToRisor(&pp), nil
}

// Cancel stops an unfinished run; the key is the entry step image's
// DeployKey.
func (p *Pipeline) Cancel(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.pipeline.cancel: expected cancel(deployKey), got %d arguments", len(args))
	}
	key, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.pipeline.cancel: deploy key must be a string (got %s)", args[0].Type())
	}
	if err := runclient.CancelPipeline(p.runner.url, p.id, key); err != nil {
		return nil, err
	}
	return object.True, nil
}

// Resume re-runs an interrupted or failed run, optionally from a named
// step: resume(deployKey) or resume(deployKey, fromStep).
func (p *Pipeline) Resume(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("tie.pipeline.resume: expected resume(deployKey[, fromStep]), got %d arguments", len(args))
	}
	key, err := object.AsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("tie.pipeline.resume: deploy key must be a string (got %s)", args[0].Type())
	}
	from := ""
	if len(args) == 2 {
		from, err = object.AsString(args[1])
		if err != nil {
			return nil, fmt.Errorf("tie.pipeline.resume: step name must be a string (got %s)", args[1].Type())
		}
	}
	pp, err := runclient.ResumePipeline(p.runner.url, p.id, key, from, nil)
	if err != nil {
		return nil, err
	}
	return NewPipeline(p.runner, pp.PipelineID), nil
}

// SubmitPipeline implements r.pipeline(req): the request map mirrors the
// frozen wire shape (docs/runner-api.md), with script-friendly step keys:
//
//	{
//	  name:  "chain",                       optional
//	  seeds: { raw: "<tree-hash>" },        optional; names a seed like a step
//	  steps: [
//	    { name: "one", image: "tool:v2", jobtype: "CPUv1",
//	      inputs: [ { from: "raw" },
//	                { from: "one", glob: "*.tsv", as: "x.tsv" },
//	                { hash: "<tree-hash>", paths: ["a", "b"] } ] },
//	  ],
//	}
//
// Inputs are selections over predecessor steps' outputs; the runner
// composes them into fresh manifests on the filehost — no bytes move.
func (r *Runner) SubmitPipeline(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.runner.pipeline: expected pipeline(requestMap), got %d arguments", len(args))
	}
	reqMap, ok := args[0].(*object.Map)
	if !ok {
		return nil, fmt.Errorf("tie.runner.pipeline: argument must be a map (got %s)", args[0].Type())
	}
	req, err := pipelineRequestFromRisor(reqMap)
	if err != nil {
		return nil, err
	}
	pp, err := runclient.SubmitPipeline(r.url, *req)
	if err != nil {
		return nil, err
	}
	return NewPipeline(r, pp.PipelineID), nil
}

func pipelineRequestFromRisor(m *object.Map) (*api.PipelineRequest, error) {
	items := m.Value()
	req := &api.PipelineRequest{}

	if v, found := items["name"]; found {
		s, err := object.AsString(v)
		if err != nil {
			return nil, fmt.Errorf("tie.runner.pipeline: name must be a string (got %s)", v.Type())
		}
		req.Name = s
	}
	if v, found := items["seeds"]; found {
		seedsMap, ok := v.(*object.Map)
		if !ok {
			return nil, fmt.Errorf("tie.runner.pipeline: seeds must be a map of step-name to tree-hash (got %s)", v.Type())
		}
		req.Seeds = map[string]string{}
		for name, hv := range seedsMap.Value() {
			h, err := object.AsString(hv)
			if err != nil {
				return nil, fmt.Errorf("tie.runner.pipeline: seeds[%q] must be a tree hash (got %s)", name, hv.Type())
			}
			req.Seeds[name] = h
		}
	}

	stepsObj, found := items["steps"]
	if !found {
		return nil, fmt.Errorf("tie.runner.pipeline: steps is required")
	}
	stepsList, ok := stepsObj.(*object.List)
	if !ok {
		return nil, fmt.Errorf("tie.runner.pipeline: steps must be a list (got %s)", stepsObj.Type())
	}
	for i, so := range stepsList.Value() {
		sm, ok := so.(*object.Map)
		if !ok {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d] must be a map (got %s)", i, so.Type())
		}
		step, err := pipelineStepFromRisor(i, sm)
		if err != nil {
			return nil, err
		}
		req.Steps = append(req.Steps, *step)
	}
	if len(req.Steps) == 0 {
		return nil, fmt.Errorf("tie.runner.pipeline: steps must not be empty")
	}
	return req, nil
}

func pipelineStepFromRisor(i int, m *object.Map) (*api.PipelineStep, error) {
	items := m.Value()
	step := &api.PipelineStep{}

	str := func(key string, required bool, dst *string) error {
		v, found := items[key]
		if !found {
			if required {
				return fmt.Errorf("tie.runner.pipeline: steps[%d].%s is required", i, key)
			}
			return nil
		}
		s, err := object.AsString(v)
		if err != nil {
			return fmt.Errorf("tie.runner.pipeline: steps[%d].%s must be a string (got %s)", i, key, v.Type())
		}
		*dst = s
		return nil
	}
	if err := str("name", true, &step.Name); err != nil {
		return nil, err
	}
	if err := str("image", true, &step.Handler); err != nil {
		return nil, err
	}
	step.JobType = "CPUv1"
	if err := str("jobtype", false, &step.JobType); err != nil {
		return nil, err
	}
	if err := str("apphash", false, &step.AppHash); err != nil {
		return nil, err
	}
	if v, found := items["settings"]; found {
		s, err := object.AsString(v)
		if err != nil {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].settings must be a JSON string (got %s)", i, v.Type())
		}
		if !json.Valid([]byte(s)) {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].settings is not valid JSON", i)
		}
		step.Settings = json.RawMessage(s)
	}
	if v, found := items["inputs"]; found {
		list, ok := v.(*object.List)
		if !ok {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].inputs must be a list (got %s)", i, v.Type())
		}
		for k, io := range list.Value() {
			im, ok := io.(*object.Map)
			if !ok {
				return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].inputs[%d] must be a map (got %s)", i, k, io.Type())
			}
			spec, err := inputSpecFromRisor(i, k, im)
			if err != nil {
				return nil, err
			}
			step.Inputs = append(step.Inputs, *spec)
		}
	}
	return step, nil
}

func inputSpecFromRisor(i, k int, m *object.Map) (*compose.Spec, error) {
	items := m.Value()
	spec := &compose.Spec{}

	if v, found := items["from"]; found {
		s, err := object.AsString(v)
		if err != nil {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].inputs[%d].from must be a string (got %s)", i, k, v.Type())
		}
		spec.FromStep = s
	}
	if v, found := items["hash"]; found {
		s, err := object.AsString(v)
		if err != nil {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].inputs[%d].hash must be a string (got %s)", i, k, v.Type())
		}
		spec.Hash = s
	}
	if spec.FromStep == "" && spec.Hash == "" {
		return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].inputs[%d] needs 'from' (a step or seed name) or 'hash' (a tree)", i, k)
	}
	if v, found := items["paths"]; found {
		paths, err := asStringList(v, fmt.Sprintf("steps[%d].inputs[%d].paths", i, k))
		if err != nil {
			return nil, err
		}
		spec.Paths = paths
	}
	if v, found := items["glob"]; found {
		s, err := object.AsString(v)
		if err != nil {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].inputs[%d].glob must be a string (got %s)", i, k, v.Type())
		}
		spec.Glob = s
	}
	if v, found := items["as"]; found {
		s, err := object.AsString(v)
		if err != nil {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].inputs[%d].as must be a string (got %s)", i, k, v.Type())
		}
		spec.As = s
	}
	if v, found := items["optional"]; found {
		b, ok := v.(*object.Bool)
		if !ok {
			return nil, fmt.Errorf("tie.runner.pipeline: steps[%d].inputs[%d].optional must be a boolean (got %s)", i, k, v.Type())
		}
		spec.Optional = b.Value()
	}
	return spec, nil
}

// pipelineToRisor / stepToRisor convert deliberately by hand rather than
// via reflection: the step record carries a json.RawMessage (Settings)
// that reflection would degrade to a byte array, and stable key names
// are the script API.
func pipelineToRisor(p *api.Pipeline) *object.Map {
	steps := make([]object.Object, 0, len(p.Steps))
	for _, s := range p.Steps {
		steps = append(steps, stepToRisor(s))
	}
	return object.NewMap(map[string]object.Object{
		"pipelineid":    object.NewString(p.PipelineID),
		"name":          object.NewString(p.Name),
		"user":          object.NewString(p.User),
		"status":        object.NewString(p.Status),
		"error_msg":     object.NewString(p.ErrorMsg),
		"debug":         object.FromGoType(p.Debug),
		"created_unix":  object.FromGoType(p.CreatedUnix),
		"finished_unix": object.FromGoType(p.FinishedUnix),
		"steps":         object.NewList(steps),
	})
}

func stepToRisor(s *api.PipelineStep) *object.Map {
	return object.NewMap(map[string]object.Object{
		"name":          object.NewString(s.Name),
		"image":         object.NewString(s.Handler),
		"jobtype":       object.NewString(s.JobType),
		"ordinal":       object.FromGoType(s.Ordinal),
		"status":        object.NewString(s.Status),
		"jobid":         object.NewString(s.JobID),
		"output_hash":   object.NewString(s.OutputHash),
		"error_msg":     object.NewString(s.ErrorMsg),
		"started_unix":  object.FromGoType(s.StartedUnix),
		"finished_unix": object.FromGoType(s.FinishedUnix),
	})
}
