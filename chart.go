package fynerisor

import (
	"context"
	"errors"
	"fmt"

	risorchart "github.com/uidbz/fynerisor/chart"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Chart{}

const ChartType object.Type = "chart"

// Chart is a factory object that provides chart creation functions to Risor scripts.
// It implements object.Object and exposes methods via GetAttr() for creating
// visualization widgets like bar charts.
//
// Available in Risor scripts as the global 'chart' object.
//
// Example usage in Risor:
//
//	let bars = chart.NewBarChart("Sales", "Revenue", ["Q1", "Q2", "Q3"], [100, 150, 200])
type Chart struct {
}

func (obj *Chart) Type() object.Type {
	return ChartType
}

func (obj *Chart) Inspect() string {
	return "chart"
}

func (obj *Chart) Interface() interface{} {
	return nil
}

func (obj *Chart) IsTruthy() bool {
	return true
}

func (obj *Chart) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'chart'")
}

func (obj *Chart) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ChartType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ChartType, opType)
	return errObj, err
}

func (obj *Chart) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Chart) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Chart) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ChartType, name)
}

func (obj *Chart) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "NewBarChart":
		return object.NewBuiltin("chart.NewBarChart", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 4 {
				return object.Errorf("wrong number of arguments. got=%d, want=4", len(args)), nil
			}
			title, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			ylabel, err := object.AsString(args[1])
			if err != nil {
				return nil, err
			}

			labels, err := object.AsStringSlice(args[2])
			if err != nil {
				return nil, err
			}

			values, err := asFloatSlice(args[3])
			if err != nil {
				return nil, err
			}

			chart, err := risorchart.NewBarChart(title, ylabel, labels, values)
			if err != nil {
				return object.NewError(err), nil
			}

			return chart, nil
		}), true
	}
	return nil, false
}

func asFloatSlice(obj object.Object) ([]float64, error) {
	list, ok := obj.(*object.List)
	if !ok {
		return nil, fmt.Errorf("type error: expected a list (%s given)", obj.Type())
	}
	result := make([]float64, 0, len(list.Value()))
	for _, item := range list.Value() {
		s, err := object.AsFloat(item)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func newChart() *Chart {
	return &Chart{}
}
