package tie

import (
	"context"
	"fmt"

	"git.sr.ht/~uid/tie/api"
	"git.sr.ht/~uid/tie/client"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const BatchType object.Type = "tie.batch"

type Batch struct {
	tc    *client.TieClient
	batch *api.Batch
}

func NewBatch(tc *client.TieClient, batch *api.Batch) *Batch {
	return &Batch{tc: tc, batch: batch}
}

func (b *Batch) Type() object.Type {
	return BatchType
}

func (b *Batch) Inspect() string {
	return "tie.batch"
}

func (b *Batch) Interface() interface{} {
	return b.batch
}

func (b *Batch) IsTruthy() bool {
	return b.batch != nil
}

func (b *Batch) Cost() int {
	return 8
}

func (b *Batch) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal tie.batch")
}

func (b *Batch) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for tie.batch: %v", opType), nil
}

func (b *Batch) Equals(other object.Object) bool {
	if other.Type() != BatchType {
		return false
	}
	return b.batch == other.(*Batch).batch
}

func (b *Batch) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: tie.batch object has no attribute %q", name)
}

func (b *Batch) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "add":
		return object.NewBuiltin("batch.add", b.Add), true
	case "delete":
		return object.NewBuiltin("batch.delete", b.Delete), true
	case "set":
		return object.NewBuiltin("batch.set", b.Set), true
	case "update":
		return object.NewBuiltin("batch.update", b.Update), true
	case "run":
		return object.NewBuiltin("batch.run", b.Run), true
	}
	return nil, false
}

func (b *Batch) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "add"},
		{Name: "delete"},
		{Name: "set"},
		{Name: "update"},
		{Name: "run"},
	}
}

func (b *Batch) Add(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("batch.add: expected 3 arguments, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	relation, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	value, err := object.AsString(args[2])
	if err != nil {
		return nil, err
	}

	b.batch.Add(key, relation, value)
	return object.Nil, nil
}

func (b *Batch) Delete(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("batch.delete: expected 3 arguments, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	relation, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	value, err := object.AsString(args[2])
	if err != nil {
		return nil, err
	}

	b.batch.Delete(key, relation, value)
	return object.Nil, nil
}

func (b *Batch) Set(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("batch.set: expected 3 arguments, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	relation, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	valuesList, ok := args[2].(*object.List)
	if !ok {
		return nil, fmt.Errorf("batch.set: third argument must be a list (got %s)", args[2].Type())
	}

	values := make([]string, 0, len(valuesList.Value()))
	for i, valObj := range valuesList.Value() {
		val, err := object.AsString(valObj)
		if err != nil {
			return nil, fmt.Errorf("batch.set: values[%d] must be a string: %w", i, err)
		}
		values = append(values, val)
	}

	b.batch.Set(key, relation, values)
	return object.Nil, nil
}

func (b *Batch) Update(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("batch.update: expected 4 arguments, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	value1, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	value2, err := object.AsString(args[2])
	if err != nil {
		return nil, err
	}

	newValue2, err := object.AsString(args[3])
	if err != nil {
		return nil, err
	}

	update := api.Update{
		Key:       key,
		Value1:    value1,
		Value2:    value2,
		NewValue2: newValue2,
	}
	b.batch.Update(update)
	return object.Nil, nil
}

func (b *Batch) Run(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("batch.run: expected 0 arguments, got %d", len(args))
	}

	_, err := b.tc.Batch(b.batch)
	if err != nil {
		return nil, err
	}

	return object.Nil, nil
}
